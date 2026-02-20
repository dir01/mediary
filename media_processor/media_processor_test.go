package media_processor

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestConcatenate_EmptyFilepathsReturnsError(t *testing.T) {
	processor := &FFMpegMediaProcessor{log: zap.NewNop()}

	_, err := processor.Concatenate(context.Background(), nil, "aac")
	if err == nil {
		t.Fatal("expected error for empty filepaths")
	}
}

func TestConcatenate_FirstFilepathWithoutExtensionReturnsError(t *testing.T) {
	processor := &FFMpegMediaProcessor{log: zap.NewNop()}

	_, err := processor.Concatenate(context.Background(), []string{"/tmp/audio"}, "aac")
	if err == nil {
		t.Fatal("expected error when first filepath has no extension")
	}
}
