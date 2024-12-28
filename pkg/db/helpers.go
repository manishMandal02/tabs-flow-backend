package db

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

func getAllStaticSKs() []string {
	return []string{
		SORT_KEY.Profile,
		SORT_KEY.Subscription,
		SORT_KEY.UsageAnalytics,
		SORT_KEY.P_General,
		SORT_KEY.P_Notes,
		SORT_KEY.P_CmdPalette,
		SORT_KEY.P_LinkPreview,
		SORT_KEY.P_AutoDiscard,
	}
}

// query dynamodb with sort key prefixes to get all dynamic sort keys
func (db *DDB) GetAllSKs(pk string) ([]string, error) {

	sortKeys := getAllStaticSKs()

	dynamicSKPrefixes := []string{
		SORT_KEY.Notifications(""),
		SORT_KEY.Space(""),
		SORT_KEY.TabsInSpace(""),
		SORT_KEY.GroupsInSpace(""),
		SORT_KEY.SnoozedTabs(""),
		SORT_KEY.Notes(""),
	}

	for _, prefix := range dynamicSKPrefixes {

		keyEx :=
			expression.Key(PK_NAME).Equal(expression.Value(pk))

		keyEx.And(expression.Key(SK_NAME).BeginsWith(prefix))

		expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()

		if err != nil {
			return sortKeys, fmt.Errorf("error building key expression for sort_key: %v", prefix)
		}

		input := &dynamodb.QueryInput{
			TableName:                 &db.TableName,
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			KeyConditionExpression:    expr.Condition(),
		}

		paginator := dynamodb.NewQueryPaginator(db.Client, input)

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(context.TODO())

			if err != nil {
				return sortKeys, fmt.Errorf("error querying for dynamic sort keys. err: %v", err)
			}

			for _, item := range page.Items {

				var sk struct {
					SK string
				}

				err := attributevalue.UnmarshalMap(item, &sk)

				if err != nil {
					return sortKeys, fmt.Errorf("error un_marshalling item for sort_key: %v", prefix)
				}

				sortKeys = append(sortKeys, sk.SK)
			}
		}
	}

	return sortKeys, nil
}

func (db *DDB) TransactionPut(keys []map[string]types.AttributeValue) error {
	var transactItems []types.TransactWriteItem

	for _, key := range keys {
		transactItems = append(transactItems, types.TransactWriteItem{
			Put: &types.Put{
				TableName:           &db.TableName,
				Item:                key,
				ConditionExpression: aws.String("attribute_not_exists(PK)"),
			},
		})
	}
	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: transactItems,
	}

	// Execute the transaction
	_, err := db.Client.TransactWriteItems(context.TODO(), input)

	if err != nil {
		return fmt.Errorf("error executing transaction [TransactionPut] [Error]: %v", err)
	}
	return nil
}

func (db *DDB) TransactionDelete(keys []map[string]types.AttributeValue) error {
	var transactItems []types.TransactWriteItem

	for _, key := range keys {
		transactItems = append(transactItems, types.TransactWriteItem{
			Delete: &types.Delete{
				TableName: &db.TableName,
				Key:       key,
			},
		})
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: transactItems,
	}

	// Execute the transaction
	_, err := db.Client.TransactWriteItems(context.TODO(), input)

	if err != nil {
		return fmt.Errorf("error executing transaction [TransactionPut] [Error]: %v", err)
	}

	return nil
}

func (db *DDB) BatchWriter(ctx context.Context, tableName string, wg *sync.WaitGroup, errChan chan error, reqs []types.WriteRequest) {

	for start := 0; start < len(reqs); start += DDB_MAX_BATCH_SIZE {
		end := start + DDB_MAX_BATCH_SIZE
		if end > len(reqs) {
			end = len(reqs)
		}

		// Prepare batch requests
		batchReqs := reqs[start:end]

		wReqs := map[string][]types.WriteRequest{}

		wReqs[tableName] = batchReqs

		wg.Add(1)

		go func(r map[string][]types.WriteRequest) {
			defer wg.Done()

			// Wait for rate limiter
			if err := db.Limiter.Wait(ctx); err != nil {
				errChan <- fmt.Errorf("rate limiter error: %w", err)
				return
			}

			// retry logic with backoff
			var lastErr error

			for attempt := 0; attempt < 5; attempt++ {
				if attempt > 0 {
					// Exponential backoff with jitter
					backoffDuration := time.Duration(math.Pow(2, float64(attempt))) * 100 * time.Millisecond
					jitter := time.Duration(rand.Float64() * float64(backoffDuration/2))
					time.Sleep(backoffDuration + jitter)
				}

				output, err := db.Client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
					RequestItems: r,
				})

				if err != nil {
					lastErr = err
					logger.Errorf("attempt %d failed: %v", attempt+1, err)
					continue
				}

				// Handle unprocessed items
				if len(output.UnprocessedItems) > 0 {
					r = output.UnprocessedItems
					lastErr = fmt.Errorf("unprocessed items remain")
					continue
				}

				// Success
				return
			}

			if lastErr != nil {
				errChan <- fmt.Errorf("batch write failed after retries: %w", lastErr)
			}
		}(wReqs)

	}

}

//! not used currently
// func (db DDB) BatchReader(ctx context.Context, wg *sync.WaitGroup, errChan chan error, keys []map[string]types.AttributeValue, res chan []map[string]types.AttributeValue) {
// 	for start := 0; start < len(keys); start += DDB_MAX_BATCH_SIZE {
// 		end := start + DDB_MAX_BATCH_SIZE
// 		if end > len(keys) {
// 			end = len(keys)
// 		}

// 		// Prepare batch requests
// 		batchReqs := keys[start:end]

// 		rReqs := map[string]types.KeysAndAttributes{}

// 		rReqs[db.TableName] = types.KeysAndAttributes{

// 			Keys: batchReqs,
// 		}

// 		wg.Add(1)

// 		go func(r map[string]types.KeysAndAttributes) {
// 			defer wg.Done()

// 			// Wait for rate limiter
// 			if err := db.Limiter.Wait(ctx); err != nil {
// 				errChan <- fmt.Errorf("rate limiter error: %w", err)
// 				return
// 			}

// 			// retry logic with backoff
// 			var lastErr error

// 			for attempt := 0; attempt < 5; attempt++ {
// 				if attempt > 0 {
// 					// Exponential backoff with jitter
// 					backoffDuration := time.Duration(math.Pow(2, float64(attempt))) * 100 * time.Millisecond
// 					jitter := time.Duration(rand.Float64() * float64(backoffDuration/2))
// 					time.Sleep(backoffDuration + jitter)
// 				}

// 				response, err := db.Client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
// 					RequestItems: r,
// 				})

// 				if err != nil {
// 					lastErr = err
// 					logger.Errorf("attempt %d failed: %v", attempt+1, err)
// 					continue
// 				}

// 				// Handle unprocessed items
// 				if len(response.UnprocessedKeys) > 0 {
// 					r = response.UnprocessedKeys
// 					lastErr = fmt.Errorf("unprocessed items remain")
// 					continue
// 				}

// 				// Success
// 				res <- response.Responses[db.TableName]
// 				return
// 			}

// 			if lastErr != nil {
// 				errChan <- fmt.Errorf("batch write failed after retries: %w", lastErr)
// 			}
// 		}(rReqs)
// 	}

// }
