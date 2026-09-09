package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/roshbhatia/go-utils/provider"
	"github.com/roshbhatia/seshy/internal/exitcode"
	"github.com/spf13/cobra"
)

// The provider/v1 capabilities seshy answers. They are the actions the
// manifest in share/seshy/providers/seshy.yaml declares.
const (
	capabilityValidate = "provider.validate"
	capabilityList     = "source.list"
	capabilityOpen     = "source.open"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Answer one provider/v1 request from stdin",
	Long: `Answer one provider/v1 request frame read from stdin, as roster invokes it.

The frame's capability selects the answer: provider.validate reports ok,
source.list returns the roster.catalog/v1 document, and source.open takes
{"id": "seshy:<name>"} and returns that row with its spawn plan. The result is
one JSON line on stdout. A capability seshy does not implement, or an id that
names no session, is an error result, not a non-zero exit; a frame that is
not a provider/v1 request is a usage error.

The manifest roster discovers is share/seshy/providers/seshy.yaml.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return serveProvider(os.Stdin, os.Stdout, time.Now())
	},
}

// serveProvider reads one request frame from in and writes one result frame
// to out. A frame that is not a provider/v1 request cannot be answered on the
// protocol, so it is returned as a usage error instead.
func serveProvider(in io.Reader, out io.Writer, now time.Time) error {
	request, err := readRequest(in)
	if err != nil {
		return exitcode.Usagef("invalid provider request: %v", err)
	}
	result := provider.Result{
		Version:   provider.Version,
		Kind:      provider.FrameResult,
		RequestID: request.RequestID,
		Status:    provider.ResultOK,
	}
	output, err := dispatch(request, now)
	if err != nil {
		result.Status = provider.ResultError
		result.Message = Message(err)
	} else {
		result.Output, err = json.Marshal(output)
		if err != nil {
			return fmt.Errorf("encode result: %w", err)
		}
	}
	return json.NewEncoder(out).Encode(result)
}

// readRequest decodes the single request frame and checks the fields the
// protocol requires before an answer can be addressed to it.
func readRequest(in io.Reader) (provider.Request, error) {
	var request provider.Request
	decoder := json.NewDecoder(in)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, err
	}
	if request.Version != provider.Version {
		return request, fmt.Errorf("version must be %q", provider.Version)
	}
	if request.Kind != provider.FrameRequest {
		return request, fmt.Errorf("kind must be %q", provider.FrameRequest)
	}
	if request.RequestID == "" {
		return request, errors.New("requestId is required")
	}
	if request.Capability == "" {
		return request, errors.New("capability is required")
	}
	return request, nil
}

// dispatch answers one capability. Its error becomes the result's message.
func dispatch(request provider.Request, now time.Time) (any, error) {
	switch request.Capability {
	case capabilityValidate:
		return map[string]bool{"ok": true}, nil
	case capabilityList:
		return catalogDocument(now)
	case capabilityOpen:
		var input struct {
			ID string `json:"id"`
		}
		if len(bytes.TrimSpace(request.Input)) > 0 {
			if err := json.Unmarshal(request.Input, &input); err != nil {
				return nil, fmt.Errorf("decode input: %w", err)
			}
		}
		if input.ID == "" {
			return nil, errors.New("input.id is required")
		}
		return openRow(input.ID)
	default:
		return nil, fmt.Errorf("unsupported capability %q", request.Capability)
	}
}

func init() {
	rootCmd.AddCommand(providerCmd)
}
