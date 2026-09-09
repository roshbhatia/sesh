// Package provider narrows the canonical provider/v1 contract to seshy's
// source manifest.
//
// The contract itself is provider.cue in roshbhatia/provider-spec, pinned as
// the provider-spec flake input. This file adds seshy's rules: the manifest
// is kind source, it answers source.list, and it declares no action outside
// #ActionName. Vet the manifest with both files:
//
//	cue vet -d '#Manifest' "$PROVIDER_SPEC/provider.cue" schema/narrow.cue share/seshy/providers/seshy.yaml
package provider

#ActionName: "provider.validate" |
	"source.list" |
	"source.open"

#Manifest: {
	kind!: "source"
	// The spec's pattern constraint admits any lowercase name, and a closed
	// struct unified with it widens rather than narrows. The hidden field
	// checks each declared name against the vocabulary instead.
	actions: [name=string]: {_supported: name & #ActionName}
	actions: "source.list"!: _
}
