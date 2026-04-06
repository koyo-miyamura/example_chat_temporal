package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func temporalClient() (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort: "temporal:7233",
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Unable to create Temporal client: %v", err))
		return nil, fmt.Errorf("unable to create Temporal client: %v", err)
	}

	return c, nil
}

const (
	ChatMessageUpdateName = "chat_update"
	ChatTaskQueue         = "CHAT_QUEUE"
)

type UserReplyInput struct {
	Text   string
	UserID string
	ChatID string
}

// ----------------------------------------------------------------------------
// Workflow
// ----------------------------------------------------------------------------
func ChatWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	processing := false

	// Single update handler for all chat messages
	var pendingInput UserReplyInput
	inputReceived := false
	if err := workflow.SetUpdateHandlerWithOptions(ctx, ChatMessageUpdateName,
		func(ctx workflow.Context, input UserReplyInput) (string, error) {
			pendingInput = input
			inputReceived = true
			return "accepted", nil
		},
		workflow.UpdateHandlerOptions{
			Validator: func(ctx workflow.Context, input UserReplyInput) error {
				if processing {
					return fmt.Errorf("currently processing, please wait")
				}
				return nil
			},
		},
	); err != nil {
		return err
	}

	// Step1: ユーザー入力を待機 → Activity実行
	if err := workflow.Await(ctx, func() bool { return inputReceived }); err != nil {
		return err
	}
	step1Input := pendingInput
	inputReceived = false
	logger.Info("Step1 started", "UserID", step1Input.UserID)

	processing = true
	var step1Result ChatStep1Result
	if err := workflow.ExecuteActivity(ctx, ChatStep1Activity, ChatStep1Input{
		UserID: step1Input.UserID,
		Data:   step1Input.Text,
	}).Get(ctx, &step1Result); err != nil {
		return err
	}
	processing = false
	logger.Info("Step1 completed", "result", step1Result.Result)

	// Step2: ユーザー入力を待機 → Activity実行
	if err := workflow.Await(ctx, func() bool { return inputReceived }); err != nil {
		return err
	}
	step2Input := pendingInput
	processing = true
	logger.Info("Step2 started", "UserID", step2Input.UserID)

	var step2Result ChatStep2Result
	if err := workflow.ExecuteActivity(ctx, ChatStep2Activity, ChatStep2Input{
		Step1Result: step1Result.Result,
		UserID:      step2Input.UserID,
		Data:        step2Input.Text,
	}).Get(ctx, &step2Result); err != nil {
		return err
	}
	logger.Info("Workflow completed", "result", step2Result.Result)

	return nil
}

// ----------------------------------------------------------------------------
// Workflow Update Handler
// ----------------------------------------------------------------------------
var (
	ErrChatAlreadyCompleted = fmt.Errorf("chat already completed")
)

func UpdateWorkflowWithUserMessage(ctx context.Context, input UserReplyInput) (string, error) {
	temporalClient, err := temporalClient()
	if err != nil {
		slog.Error(fmt.Sprintf("Unable to create client: %v", err))
		os.Exit(1)
	}
	defer temporalClient.Close()

	updateHandle, err := temporalClient.UpdateWithStartWorkflow(
		ctx,
		client.UpdateWithStartWorkflowOptions{
			StartWorkflowOperation: temporalClient.NewWithStartWorkflowOperation(
				client.StartWorkflowOptions{
					ID:                       input.ChatID,
					TaskQueue:                ChatTaskQueue,
					WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
					WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
				},
				ChatWorkflow,
			),
			UpdateOptions: client.UpdateWorkflowOptions{
				WorkflowID:   input.ChatID,
				UpdateName:   ChatMessageUpdateName,
				Args:         []any{input},
				WaitForStage: client.WorkflowUpdateStageCompleted,
			},
		},
	)
	if err != nil {
		if temporal.IsWorkflowExecutionAlreadyStartedError(err) {
			return "", ErrChatAlreadyCompleted
		}
		slog.Error(fmt.Sprintf("[handleChat] UpdateWithStart failed: %v", err))
		return "", fmt.Errorf("failed: %v", err)
	}

	var result string
	if err := updateHandle.Get(ctx, &result); err != nil {
		slog.Error(fmt.Sprintf("[handleChat] Update error: %v", err))
		return "", fmt.Errorf("update error: %v", err)
	}

	return result, nil
}

// ----------------------------------------------------------------------------
// Activities
// ----------------------------------------------------------------------------
type ChatStep1Input struct {
	UserID string
	Data   string
}

type ChatStep1Result struct {
	Result string
}

func ChatStep1Activity(ctx context.Context, input ChatStep1Input) (ChatStep1Result, error) {
	time.Sleep(3 * time.Second)
	return ChatStep1Result{Result: "Step1Result: " + input.Data}, nil
}

type ChatStep2Input struct {
	Step1Result string
	UserID      string
	Data        string
}

type ChatStep2Result struct {
	Result string
}

func ChatStep2Activity(ctx context.Context, input ChatStep2Input) (ChatStep2Result, error) {
	time.Sleep(3 * time.Second)
	return ChatStep2Result{Result: "Step2Result: " + input.Step1Result}, nil
}
