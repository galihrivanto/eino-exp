package main

import (
	"context"
	"log"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func createTemplate() prompt.ChatTemplate {
	// 创建模板，使用 FString 格式
	return prompt.FromMessages(schema.FString,
		// 系统消息模板
		schema.SystemMessage("You are a {role}. You need to answer questions in a {style} tone. Your goal is to help programmers stay positive and optimistic, and pay attention to their mental health while providing technical advice."),

		// 插入需要的对话历史（新对话的话这里不填）
		schema.MessagesPlaceholder("chat_history", true),

		// 用户消息模板
		schema.UserMessage("question: {question}"),
	)
}

func createMessagesFromTemplate() []*schema.Message {
	template := createTemplate()

	// 使用模板生成消息
	messages, err := template.Format(context.Background(), map[string]any{
		"role":     "Programmer Encouragement",
		"style":    "positive, warm, and professional",
		"question": "My code is always reporting errors, and I feel very frustrated, what should I do?",
		// 对话历史（这个例子里模拟两轮对话历史）
		"chat_history": []*schema.Message{
			schema.UserMessage("Hello"),
			schema.AssistantMessage("Hey! I'm your programmer encouragement! Remember, every great programmer grows from debugging. What can I help you with?", nil),
			schema.UserMessage("I feel like my code is too bad"),
			schema.AssistantMessage("Every programmer has gone through this stage! What's important is that you're constantly learning and improving. Let's look at the code together. I believe it will get better through refactoring and optimization. Remember, Rome wasn't built in a day, and code quality is improved through continuous improvement.", nil),
		},
	})
	if err != nil {
		log.Fatalf("format template failed: %v\n", err)
	}
	return messages
}
