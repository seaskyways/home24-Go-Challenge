package webalayzer

import (
	"context"
	"os"
	"testing"
)

func TestAnalyzer_AnalyzeReader(t *testing.T) {
	testPage, _ := os.Open("../templates/test.gohtml")
	defer testPage.Close()

	analyzer := new(Analyzer)
	webPageStats, err := analyzer.AnalyzeReader(context.WithValue(context.Background(), "url", "http://localhost:8080/test"), testPage)

	if err != nil {
		t.Fatalf("error: %v", err)
	} else {
		t.Logf("stats: %#v", webPageStats)
	}
}
