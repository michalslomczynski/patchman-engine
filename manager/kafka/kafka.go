package kafka

import (
	"app/base"
	"app/base/database"
	"app/base/models"
	"app/base/mqueue"
	"app/base/utils"
)

var (
	evalWriter               mqueue.Writer
	inventoryIDsChan         chan inventoryIDsBatch
	enableBaselineChangeEval = utils.GetBoolEnvOrDefault("ENABLE_BASELINE_CHANGE_EVAL", true)
)

type inventoryIDsBatch struct {
	InventoryIDs []mqueue.EvalData
}

func TryStartEvalQueue(createWriter mqueue.CreateWriter) {
	if !enableBaselineChangeEval {
		return
	}
	evalTopic := utils.FailIfEmpty(utils.Cfg.EvalTopic, "EVAL_TOPIC")
	evalWriter = createWriter(evalTopic)
	inventoryIDsChan = make(chan inventoryIDsBatch)
	go runBaselineRecalcLoop()
}

func runBaselineRecalcLoop() {
	for {
		batch := <-inventoryIDsChan
		sendInventoryIDs(batch.InventoryIDs)
	}
}

func GetInventoryIDsToEvaluate(baselineID *int, accountID int,
	configUpdated bool, updatedInventoryIDs []string) []mqueue.EvalData {
	if !enableBaselineChangeEval {
		return nil
	}

	if !configUpdated && updatedInventoryIDs == nil {
		return nil // no evaluation needed for no config and inventory IDs updates
	}

	var inventoryAIDs []mqueue.EvalData
	if !configUpdated { // we just need to evaluate updated inventory IDs
		inventoryAIDs = inventoryIDs2InventoryAIDs(accountID, updatedInventoryIDs)
	} else { // config updated - we need to update all baseline inventory IDs and the added ones too
		inventoryAIDs = getInventoryIDs(baselineID, accountID, updatedInventoryIDs)
	}

	utils.Log("nInventoryIDs", len(inventoryAIDs), "accountID", accountID).
		Debug("Loaded inventory IDs to evaluate")
	return inventoryAIDs
}

func inventoryIDs2InventoryAIDs(accountID int, inventoryIDs []string) []mqueue.EvalData {
	inventoryAIDs := make([]mqueue.EvalData, 0, len(inventoryIDs))
	for _, v := range inventoryIDs {
		inventoryAIDs = append(inventoryAIDs, mqueue.EvalData{InventoryID: v, RhAccountID: accountID})
	}
	return inventoryAIDs
}

func getInventoryIDs(baselineID *int, accountID int, inventoryIDs []string) []mqueue.EvalData {
	var inventoryAIDs []mqueue.EvalData
	query := database.Db.Model(&models.SystemPlatform{}).
		Select("inventory_id, rh_account_id").
		Where(map[string]interface{}{"rh_account_id": accountID, "baseline_id": baselineID})

	if len(inventoryIDs) > 0 {
		query = query.Or("inventory_id IN (?) AND rh_account_id = ?", inventoryIDs, accountID)
	}

	err := query.Order("inventory_id").
		Scan(&inventoryAIDs).Error
	if err != nil {
		utils.Log("err", err.Error()).
			Error("Unable to load inventory IDs for baseline")
	}
	return inventoryAIDs
}

func sendInventoryIDs(inventoryIDs mqueue.EvalDataSlice) {
	if len(inventoryIDs) == 0 {
		return
	}

	err := mqueue.SendMessages(base.Context, evalWriter, &inventoryIDs)
	if err != nil {
		utils.Log("nInventoryIDs", len(inventoryIDs), "err", err.Error()).
			Error("Inventory IDs sending failed")
	}
}

// Send all account systems of given baseline to evaluation.
// Evaluate all account systems with no baseline if baselineID is nil (used for deleted baseline).
func EvaluateBaselineSystems(inventoryAIDs []mqueue.EvalData) {
	if !enableBaselineChangeEval {
		return
	}

	batch := inventoryIDsBatch{InventoryIDs: inventoryAIDs}
	inventoryIDsChan <- batch
}
