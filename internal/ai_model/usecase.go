package ai_model

import (
	"context"
	"log"
)

type UseCaseMultiple struct {
	Models []AiModel
}

func NewMultUseCase(models []AiModel) UseCaseMultiple {
	return UseCaseMultiple{models}
}

func (repo *UseCaseMultiple) AskMultiple(ctx context.Context, chatId int64, inputFor InputForm, onLoad func(reply string)) {
	log.Println("[HuggingFace.AskMultiple model: ", repo.Models)
	for _, model := range repo.Models {
		log.Println("[UseCaseMultiple.AskMultiple] ask: ", model)
		reply := model.AskGpt(ctx, chatId, inputFor)
		onLoad(reply)
	}
}
