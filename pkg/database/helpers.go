package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// types
type getAllSKsType func(db *DDB, pk string) ([]string, error)

// the only exported variable in this file
var Helpers = struct {
	GetAllSKs getAllSKsType
}{
	GetAllSKs: getAllSKs,
}

// query dynamodb with sort key prefixes to get all dynamic sort keys
func getAllSKs(db *DDB, pk string) ([]string, error) {
	sortKeys := getAllStaticSKs()

	dynamicSKPrefixes := []string{
		SORT_KEY.Notifications(""),
		SORT_KEY.Space(""),
		SORT_KEY.TabsInSpace(""),
		SORT_KEY.GroupsInSpace(""),
		SORT_KEY.SnoozedTabs(""),
		SORT_KEY.Note(""),
	}

	for _, prefix := range dynamicSKPrefixes {

		keyEx :=
			expression.Key(PK_NAME).Equal(expression.Value(pk))

		keyEx.And(expression.Key(SK_NAME).BeginsWith(prefix))

		expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()

		if err != nil {
			return nil, fmt.Errorf("error building key expression for sort_key: %v", prefix)
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
				return nil, fmt.Errorf("error querying for dynamic sort keys. err: %v", err)
			}

			for _, item := range page.Items {

				var sk struct {
					SK string
				}

				err := attributevalue.UnmarshalMap(item, &sk)

				if err != nil {
					return nil, fmt.Errorf("error un_marshalling item for sort_key: %v", prefix)
				}

				sortKeys = append(sortKeys, sk.SK)
			}
		}
	}

	return []string{}, nil
}

func getAllStaticSKs() []string {
	return []string{
		SORT_KEY.Profile,
		SORT_KEY.Subscription,
		SORT_KEY.UsageAnalytics,
		SORT_KEY.P_General,
		SORT_KEY.P_Note,
		SORT_KEY.P_CmdPalette,
		SORT_KEY.P_LinkPreview,
		SORT_KEY.P_AutoDiscard,
	}
}
