package domain

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/kai5263499/rhema/generated"
	v1 "github.com/kai5263499/rhema/internal/v1"
)

func ConvertParamsToProto(submitRequests *v1.SubmitRequestJSONRequestBody) (requests []*generated.Request) {
	requests = make([]*generated.Request, len(*submitRequests))

	for idx, submitRequest := range *submitRequests {
		requests[idx] = &generated.Request{
			RequestHash: uuid.Must(uuid.NewV4()).String(),
			Uri:         submitRequest.Uri,
			SubmittedAt: uint64(time.Now().UTC().Unix()),
			Created:     uint64(time.Now().UTC().Unix()),
		}

		if submitRequest.Text != nil {
			requests[idx].Text = *submitRequest.Text
		}
		if submitRequest.Title != nil {
			requests[idx].Title = *submitRequest.Title
		}
		if submitRequest.EspeakVoice != nil {
			requests[idx].ESpeakVoice = *submitRequest.EspeakVoice
		}
		if submitRequest.Atempo != nil {
			requests[idx].ATempo = *submitRequest.Atempo
		}
		if submitRequest.WordsPerMinute != nil {
			requests[idx].WordsPerMinute = *submitRequest.WordsPerMinute
		}
		if submitRequest.SubmittedBy != nil {
			requests[idx].SubmittedBy = *submitRequest.SubmittedBy
		}
	}

	return
}

func ConvertProtoToInputParams(r *generated.Request) (o *v1.SubmitRequestInput) {
	contentType := r.Type.String()
	o = &v1.SubmitRequestInput{
		Uri:                 r.Uri,
		RequestHash:         &r.RequestHash,
		Title:               &r.Title,
		Atempo:              &r.ATempo,
		WordsPerMinute:      &r.WordsPerMinute,
		Length:              &r.Length,
		Size:                &r.Size,
		Text:                &r.Text,
		NumberOfConversions: &r.NumberOfConversions,
		Type:                &contentType,
		Created:             &r.Created,
	}

	if r.SubmittedAt > 0 {
		submittedAt := int(r.SubmittedAt)
		o.SubmittedAt = &submittedAt
	}

	return
}

func ConvertProtoToOutputParams(r *generated.Request) (o *v1.SubmitRequestOutput) {

	contentType := r.Type.String()
	submittedAt := 0
	if r.SubmittedAt > 0 {
		submittedAt = int(r.SubmittedAt)
	}

	o = &v1.SubmitRequestOutput{
		{
			Uri:                 r.Uri,
			RequestHash:         &r.RequestHash,
			Title:               &r.Title,
			Atempo:              &r.ATempo,
			WordsPerMinute:      &r.WordsPerMinute,
			Length:              &r.Length,
			Size:                &r.Size,
			Text:                &r.Text,
			NumberOfConversions: &r.NumberOfConversions,
			Type:                &contentType,
			Created:             &r.Created,
			SubmittedAt:         &submittedAt,
		},
	}

	return
}
