package main

import (
  "context"
  "errors"
  "fmt"

  "github.com/sashabaranov/go-openai"
)

// gptGPT3CompleteContext: Call GPT3.5 API
func gptGPT3CompleteContext(ori string) (ret string) {
  fmt.Printf("Using GPT3 Complete")
  return gptCompleteContext(ori, openai.GPT3Dot5Turbo)
}

// gptGPT4CompleteContext: Call GPT4 API
func gptGPT4CompleteContext(ori string) (ret string) {
  fmt.Printf("Using GPT4 Complete")
  return gptCompleteContext(ori, openai.GPT4)
}

func gptCompleteContext(ori string, model string) (ret string) {
  // Get the context.
  ctx := context.Background()

  // For more details about the API of Open AI Chat Completion: https://platform.openai.com/docs/guides/chat
  req := openai.ChatCompletionRequest{
    Model: model,
    // The message to complete.
    Messages: []openai.ChatCompletionMessage{{
      Role:    openai.ChatMessageRoleUser,
      Content: ori,
    }},
  }

  resp, err := client.CreateChatCompletion(ctx, req)
  if err != nil {
    ret = fmt.Sprintf("Err: %v", err)
  } else {
    // The response contains a list of choices, each with a score.
    // The score is a float between 0 and 1, with 1 being the most likely.
    // The choices are sorted by score, with the first choice being the most likely.
    // So we just take the first choice.
    ret = resp.Choices[0].Message.Content
  }

  return ret
}

func gptChat(ori string, systemMessage string) (ret string, err error) {
  // Get the context.
  ctx := context.Background()

    // Create a slice of ChatCompletionMessage to construct the conversation.
  conversation := []openai.ChatCompletionMessage{
    {
      Role:    openai.ChatMessageRoleSystem,
      Content: systemMessage,
    },
    {
      Role:    openai.ChatMessageRoleUser,
      Content: ori,
    },
  }
  
  // 在傳送Chat Conversation之前，将Chat Conversation印出到log中
  fmt.Println("Chat Conversation:")
  for _, msg := range conversation {
      fmt.Printf("Role: %s, Content: %s\n", msg.Role, msg.Content)
  }

  // Create a ChatCompletionRequest using the conversation.
  req := openai.ChatCompletionRequest{
    Model:    "gpt-3.5-turbo-16k-0613",
    Messages: conversation,
    Temperature: 0.2,
  }

  resp, err := client.CreateChatCompletion(ctx, req)
  if err != nil {
    return "", err
  }

  // Extract the content from the response.
  ret = resp.Choices[0].Message.Content
  return ret, nil
}


// Create image by DALL-E 2
func gptImageCreate(prompt string) (string, error) {
  ctx := context.Background()

  // Sample image by link
  reqUrl := openai.ImageRequest{
    Prompt:         prompt,
    Size:           openai.CreateImageSize512x512,
    ResponseFormat: openai.CreateImageResponseFormatURL,
    N:              1,
  }

  respUrl, err := client.CreateImage(ctx, reqUrl)
  if err != nil {
    fmt.Printf("Image creation error: %v\n", err)
    return "", errors.New("Image creation error")
  }
  fmt.Println(respUrl.Data[0].URL)

  return respUrl.Data[0].URL, nil
}