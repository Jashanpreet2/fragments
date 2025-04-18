package fragment

import (
	"context"
	"os"

	"github.com/Jashanpreet2/fragments/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type FragmentsDynamoDBClient struct {
	ddbClient *dynamodb.Client
	TableName string
}

var fragmentsDynamoDBClient *FragmentsDynamoDBClient

func GetDynamoDBClient() (*FragmentsDynamoDBClient, error) {
	if fragmentsDynamoDBClient != nil {
		return fragmentsDynamoDBClient, nil
	}
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	ddbClient := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.Region = "us-east-1"
	})
	fragmentsDynamoDBClient = &FragmentsDynamoDBClient{ddbClient, "fragments"}
	return fragmentsDynamoDBClient, nil
}

func (fragmentsClient *FragmentsDynamoDBClient) WriteFragment(frag *Fragment) error {
	logger.Sugar.Info(frag.GetJson())
	item, err := attributevalue.MarshalMap(frag)
	var outFrag Fragment
	attributevalue.UnmarshalMap(item, &outFrag)
	data, _ := outFrag.GetJson()
	logger.Sugar.Info("Fragment that was stored:", data)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(fragmentsClient.TableName),
		Item:      item,
	}
	logger.Sugar.Info("Aws access key value" + os.Getenv("aws_access_key_id"))
	logger.Sugar.Info("Aws cognito env value" + os.Getenv("AWS_COGNITO_POOL_ID"))
	_, err = fragmentsClient.ddbClient.PutItem(context.Background(), input)
	logger.Sugar.Info(err)
	return err
}

// Returns:
//
//	nil, nil: No errors occured by a matching value was not found.
func (fragmentsClient *FragmentsDynamoDBClient) GetFragment(ownerId string, id string) (*Fragment, error) {

	logger.Sugar.Info("dynamogetdetails: ", ownerId, " ", id)
	input := &dynamodb.GetItemInput{
		TableName: aws.String(fragmentsClient.TableName),
		Key: map[string]types.AttributeValue{
			"ownerId": &types.AttributeValueMemberS{Value: ownerId},
			"id":      &types.AttributeValueMemberS{Value: id},
		},
	}
	logger.Sugar.Info(input.AttributesToGet)

	var frag Fragment
	out, err := fragmentsClient.ddbClient.GetItem(context.TODO(), input)

	if len(out.Item) == 0 {
		return nil, nil
	}
	attributevalue.UnmarshalMap(out.Item, &frag)
	logger.Sugar.Info(err)
	return &frag, err
}

func (fragmentsClient *FragmentsDynamoDBClient) GetFragmentIds(ownerId string) ([]string, error) {
	out, err := fragmentsClient.ddbClient.Query(context.TODO(), &dynamodb.QueryInput{
		TableName: aws.String(fragmentsClient.TableName),
		KeyConditions: map[string]types.Condition{
			"ownerId": {
				ComparisonOperator: types.ComparisonOperatorEq,

				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: ownerId},
				},
			},
		},

		ProjectionExpression: aws.String("id"),
	})

	if err != nil {
		return nil, err
	}

	ids := []string{}
	for _, item := range out.Items {
		if id, ok := item["id"].(*types.AttributeValueMemberS); ok {
			ids = append(ids, id.Value)
		}
	}

	return ids, nil
}

func (fragmentsClient *FragmentsDynamoDBClient) deleteFragment(ownerId string, fragmentId string) error {
	_, err := fragmentsClient.ddbClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(fragmentsClient.TableName),
		Key: map[string]types.AttributeValue{
			"ownerId": &types.AttributeValueMemberS{Value: ownerId},
			"id":      &types.AttributeValueMemberS{Value: fragmentId},
		},
	})
	if err != nil {
		logger.Sugar.Error(err)
		return err
	}
	return nil
}
