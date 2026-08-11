package envelope_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sraphaz/arah-harness/internal/envelope"
)

func TestOKAndFailEnvelope(t *testing.T) {
	ok := envelope.OK(map[string]any{"x": 1})
	if !ok.OK || ok.Code != envelope.CodeOK || ok.TraceID == "" {
		t.Fatalf("%+v", ok)
	}
	fail := envelope.Fail(envelope.CodeUsage, "bad", map[string]any{"k": "v"}, "try help")
	if fail.OK || fail.Code != envelope.CodeUsage || len(fail.Remediation) != 1 {
		t.Fatalf("%+v", fail)
	}
	var buf bytes.Buffer
	if code := envelope.WriteJSON(&buf, fail); code != 10 {
		t.Fatalf("exit=%d", code)
	}
	var parsed envelope.Envelope
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Message != "bad" {
		t.Fatalf("message=%s", parsed.Message)
	}
}
