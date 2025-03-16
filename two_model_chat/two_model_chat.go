/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	callbacks2 "github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino/utils/callbacks"

	"github.com/galihrivanto/eino-exp/internal/logs"
)

func main() {
	ctx := context.Background()

	type state struct {
		currentRound int
		msgs         []*schema.Message
	}

	llm, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434", // Ollama 服务地址
		Model:   "deepseek-r1:latest",     // 模型名称
	})
	if err != nil {
		log.Fatalf("create ollama chat model failed: %v", err)
	}

	g := compose.NewGraph[[]*schema.Message, *schema.Message](compose.WithGenLocalState(func(ctx context.Context) *state { return &state{} }))
	_ = g.AddChatModelNode("writer", llm, compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *state) ([]*schema.Message, error) {
		state.currentRound++
		state.msgs = append(state.msgs, input...)
		input = append([]*schema.Message{schema.SystemMessage("you are a writer who writes jokes and revise it according to the critic's feedback. Prepend your joke with your name which is \"writer: \"")}, state.msgs...)
		return input, nil
	}), compose.WithNodeName("writer"))
	_ = g.AddChatModelNode("critic", llm, compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *state) ([]*schema.Message, error) {
		state.msgs = append(state.msgs, input...)
		input = append([]*schema.Message{schema.SystemMessage("you are a critic who ONLY gives feedback about jokes, emphasizing on funniness. Prepend your feedback with your name which is \"critic: \"")}, state.msgs...)
		return input, nil
	}), compose.WithNodeName("critic"))
	_ = g.AddLambdaNode("toList1", compose.ToList[*schema.Message]())
	_ = g.AddLambdaNode("toList2", compose.ToList[*schema.Message]())

	_ = g.AddEdge(compose.START, "writer")
	_ = g.AddBranch("writer", compose.NewStreamGraphBranch(func(ctx context.Context, input *schema.StreamReader[*schema.Message]) (string, error) {
		input.Close()

		var result string = "toList1"
		err := compose.ProcessState(ctx, func(ctx context.Context, state *state) error {
			if state.currentRound >= 3 {
				result = compose.END
			}

			return nil
		})

		return result, err
	}, map[string]bool{compose.END: true, "toList1": true}))
	_ = g.AddEdge("toList1", "critic")
	_ = g.AddEdge("critic", "toList2")
	_ = g.AddEdge("toList2", "writer")

	runner, err := g.Compile(ctx)
	if err != nil {
		logs.Fatalf("compile error: %v", err)
	}

	sResponse := &streamResponse{
		ch: make(chan string),
	}
	go func() {
		for m := range sResponse.ch {
			fmt.Print(m)
		}
	}()
	handler := callbacks.NewHandlerHelper().ChatModel(&callbacks.ModelCallbackHandler{
		OnEndWithStreamOutput: sResponse.OnStreamStart,
	}).Handler()

	outStream, err := runner.Stream(ctx, []*schema.Message{schema.UserMessage("write a funny line about former indonesian president joko widodo, in 20 words.")},
		compose.WithCallbacks(handler))
	if err != nil {
		logs.Fatalf("stream error: %v", err)
	}
	for {
		_, err := outStream.Recv()
		if err == io.EOF {
			close(sResponse.ch)
			break
		}
	}
}

type streamResponse struct {
	ch chan string
}

func (s *streamResponse) OnStreamStart(ctx context.Context, runInfo *callbacks2.RunInfo, input *schema.StreamReader[*model.CallbackOutput]) context.Context {
	defer input.Close()
	s.ch <- "\n=======\n"
	for {
		frame, err := input.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			logs.Fatalf("internal error: %s\n", err)
		}

		s.ch <- frame.Message.Content
	}
	return ctx
}
