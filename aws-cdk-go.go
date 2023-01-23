package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsappsync"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("AwsCdkGo"), &awscdk.StackProps{})

	//認証用のCogintoユーザープールの作成
	awscognito.NewUserPool(stack, jsii.String("UserPool"), &awscognito.UserPoolProps{
		UserPoolName: jsii.String("app-userpool"),
	})

	api := awsappsync.NewCfnGraphQLApi(stack, jsii.String("BooksApi"), &awsappsync.CfnGraphQLApiProps{
		Name:               jsii.String("books-api"),
		AuthenticationType: jsii.String("API_KEY"),
	})
	awsappsync.NewCfnApiKey(stack, jsii.String("BooksApiKey"), &awsappsync.CfnApiKeyProps{
		ApiId: api.AttrApiId(),
	})
	table := awsdynamodb.NewTable(stack, jsii.String("demo-table"), &awsdynamodb.TableProps{
		TableName: jsii.String("booktable"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode: awsdynamodb.BillingMode_PAY_PER_REQUEST,
		Stream:      awsdynamodb.StreamViewType_NEW_IMAGE,
	})
	tablerole := awsiam.NewRole(stack, jsii.String("dynamodb-role"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("appsync.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
	})
	tablerole.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonDynamoDBFullAccess")))
	Ds := awsappsync.NewCfnDataSource(stack, jsii.String("DataStore"), &awsappsync.CfnDataSourceProps{
		ApiId: api.AttrApiId(),
		Name:  jsii.String("BookDataSource"),
		Type:  jsii.String("AMAZON_DYNAMODB"),
		DynamoDbConfig: awsappsync.CfnDataSource_DynamoDBConfigProperty{
			TableName: table.TableName(),
			AwsRegion: stack.Region(),
		},
		ServiceRoleArn: tablerole.RoleArn(),
	})
	def, err := os.ReadFile(filepath.Join(".", "resource", "schema.graphql"))
	if err != nil {
		fmt.Println("failed to load graphql definition " + err.Error())
	}

	schema := awsappsync.NewCfnGraphQLSchema(stack, jsii.String("GraphSchema"), &awsappsync.CfnGraphQLSchemaProps{
		ApiId:      api.AttrApiId(),
		Definition: jsii.String(string(def)),
	})

	getitem, err := os.ReadFile(filepath.Join(".", "resource", "getitem.vtl"))
	if err != nil {
		fmt.Println("failed to load  getitem.vtl " + err.Error())
	}
	putitem, err := os.ReadFile(filepath.Join(".", "resource", "putitem.vtl"))
	if err != nil {
		fmt.Println("failed to load  putitem.vtl " + err.Error())
	}

	awsappsync.NewCfnResolver(stack, jsii.String("GetResolver"), &awsappsync.CfnResolverProps{
		ApiId:                   api.AttrApiId(),
		TypeName:                jsii.String("Query"),
		FieldName:               jsii.String("getPost"),
		DataSourceName:          Ds.Name(),
		RequestMappingTemplate:  jsii.String(string(getitem)),
		ResponseMappingTemplate: jsii.String(`$util.toJson($ctx.result)`),
	}).AddDependency(schema)

	awsappsync.NewCfnResolver(stack, jsii.String("AddResolver"), &awsappsync.CfnResolverProps{
		ApiId:                   api.AttrApiId(),
		TypeName:                jsii.String("Mutation"),
		FieldName:               jsii.String("addPost"),
		DataSourceName:          Ds.Name(),
		RequestMappingTemplate:  jsii.String(string(putitem)),
		ResponseMappingTemplate: jsii.String(`$util.toJson($ctx.result)`),
	}).AddDependency(schema)
	app.Synth(nil)

}
