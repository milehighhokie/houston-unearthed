package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// Message - Define base message object
type Message struct {
	Description string `json:"Description"`
	Choice1     string `json:"Choice1"`
	Choice2     string `json:"Choice2"`
	Choice3     string `json:"Choice3"`
}

// PlayerInfo - basic player information
type PlayerInfo struct {
	Email      string `json:"Email"`
	FirstName  string `json:"FirstName"`
	LastName   string `json:"LastName"`
	Phone      string `json:"Phone"`
	PlayerName string `json:"PlayerName"`
}

// Decision - the choice made and a reason
type Decision struct {
	Choice    string `json:"Choice"`
	Reason    string `json:"Reason"`
	Task      string `json:"Task"`
	TimeTaken string `json:"Time taken"`
}

// Score - should be a slice, but it isn't
type Score struct {
	Customer string `json:"Customer Satisfaction"`
	Profit   string `json:"Profit"`
	Safety   string `json:"Safety"`
}

// Result - Collection of decisions - result of a game
type Result struct {
	DatePlayed string     `json:"DatePlayed"`
	Decisions  []Decision `json:"Decisions"`
	GameName   string     `json:"GameName"`
	Scores     Score      `json:"Scores"`
}

// Players - All the players who have registered
type Players struct {
	PlayerID   string     `json:"PlayerID"`
	PlayerInfo PlayerInfo `json:"PlayerInfo"`
	Results    []Result   `json:"Results"`
}

// GamesInProgress - All the games currently running
type GamesInProgress struct {
	GIPID     string `json:"GIPID"`
	Scheduled string `json:"Scheduled"`
	PlayerID  string `json:"PlayerID"`
}

func main() {
	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)

	if err != nil {
		fmt.Println("Got error creating session:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	svc := dynamodb.New(sess)

	// listPlayers(svc)

	go startNewGames(svc)

	go manageQueues(svc)

	//go scoreGames(svc)

	fmt.Println("Close program?")
	var input string
	fmt.Scanln(&input)
}

func startNewGames(svc *dynamodb.DynamoDB) {
	// keep looping forever checking for schedules that need to be set up
	// Once set up, change Scheduled to Y
Forever:
	for {
		// Query all GamesInProgress entries where Scheduled = N
		filt := expression.Name("Scheduled").Equal(expression.Value("N"))
		proj := expression.NamesList(expression.Name("title"), expression.Name("year"), expression.Name("info.rating"))

		expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
		if err != nil {
			fmt.Println(err)
		}
		params := &dynamodb.ScanInput{
			TableName:                 aws.String("GamesInProgress"),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
			ProjectionExpression:      expr.Projection(),
		}

		// Make the DynamoDB Query API call
		result, err := svc.Scan(params)
		if err != nil {
			fmt.Println("Got error querying:")
			fmt.Println(err.Error())
			//os.Exit(1)
		}

		gamesInProgress := []GamesInProgress{}
		err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &gamesInProgress)
		if err != nil {
			fmt.Println("Got error unmarshalling:")
			fmt.Println(err.Error())
			//	os.Exit(1)
		}

		for i, gameInProgress := range gamesInProgress {
			fmt.Println("GIP #", i, " ", gameInProgress)
			if gameInProgress.Scheduled == "Y" {
				break Forever
			} else {
				fmt.Println("need to code schedule method")
			}
		}
	}
}

func manageQueues(svcDDB *dynamodb.DynamoDB) {
	//var queueURL = "https://sqs.us-west-2.amazonaws.com/134237506622/lsb_queue.fifo"
	//var queueName = "lsb_queue.fifo"
	var queueURL = "https://sqs.us-west-2.amazonaws.com/134237506622/LSB_standard_queue"
	//var queueName = "LSB_standard_queue"

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)
	// sess := session.Must(session.NewSessionWithOptions(session.Options{
	// 	SharedConfigState: session.SharedConfigEnable,
	// }))
	Colors := []string{"lightgreen", "red", "yellow", "green", "darkorange", "yellowgreen"}
	Choice1 := []string{"Yes", "No", "Maybe", "Ask me again later", "I do not know", "Tell me more"}
	Choice2 := []string{"Turn on oven", "Turn off oven", "Turn on stove", "Turn off stove", "Put another shrimp on the barbie", "Add fuel to the fire"}
	Choice3 := []string{"Fire them!", "Hire them!!", "Go for it!", "I am not so sure...", "Really? Those are my only choices?"}
	FBIDs := []string{"10156224228933750", "1479943365447346", "134020007416753"}
	truths := []string{"cats are evil.",
		"all squirrles are Communists.",
		"goldfish are spies for The Man.",
		"selfies are the start of the downfall of civilation.",
		"television is good brain food.",
		"carrots build strong incus.",
		"Foo Fighters fight foo famously.",
		"Tinky Winky's pinky is stinky.",
		"Pluto is a planet.",
		"Le Target is A-OK.",
		"The A-Team should be more widely accepted as great art.",
		"taglinarini is a super food.",
		"otters secretly fund political action committees.",
		"better to have the Earth be flat than your soda.",
		"clouds are a government conspiracy to hide alien technology.",
		"Yeti are cuddly.",
		"The Loch Ness monster is really an unclassfied submarine.",
		"hippos are fast in the water.",
		"aspen trees are really uppity.",
		"sea horses are telepathic.",
		"unicorns exist but are invisible.",
		"violinists cannot wear earmuffs.",
		"nimble walruses are good at tamburello.",
		"quidditch should be an Olympic sport.",
		"Sepak takraw is not a Vulcan curse word.",
		"unmoist noodles are never a good conversation topic.",
	}

	svcQ := sqs.New(sess)
	for y := 0; y < 13; y++ {

		for j := 0; j < 3; j++ {
			for i := 0; i < 8; i++ {
				r := rand.Intn(len(truths))
				r3 := rand.Intn(len(Colors))
				rc1 := rand.Intn(len(Choice1))
				rc2 := rand.Intn(len(Choice2))
				rc3 := rand.Intn(len(Choice3))

				msg := Message{truths[r], Choice1[rc1], Choice2[rc2], Choice3[rc3]}
				msgJSON, _ := json.Marshal(msg)

				result, err := svcQ.SendMessage(&sqs.SendMessageInput{
					//DelaySeconds: aws.Int64(10),
					MessageAttributes: map[string]*sqs.MessageAttributeValue{
						"PlayerID": &sqs.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(FBIDs[j]),
						},
						"ScreenID": &sqs.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(strconv.Itoa(i + 1)),
						},
						"ScreenColor": &sqs.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(Colors[r3]),
						},
					},
					MessageBody: aws.String(string(msgJSON)),
					QueueUrl:    &queueURL,
					// MessageGroupId: &queueName,
					//	MessageDeduplicationId: aws.String(strconv.Itoa(i + time.Now().Second())),
				})

				if err != nil {
					fmt.Println("Error", err)
					return
				}

				fmt.Println("Success", *result.MessageId)
			}
		}
	}
}

func listPlayers(svc *dynamodb.DynamoDB) {

	// Scan all Players entries
	params := &dynamodb.ScanInput{
		TableName: aws.String("Players"),
	}

	// Make the DynamoDB Scan API call
	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println("Got error scanning:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	players := []Players{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &players)
	if err != nil {
		fmt.Println("Got error unmarshalling:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for i, player := range players {
		fmt.Println("Player #", i, " ", player)
	}
}
