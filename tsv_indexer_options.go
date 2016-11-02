package iosupport

type Options struct {
	Header                 bool
	Separator              byte
	Fields                 []string
	DropEmptyIndexedFields bool
	SkipMalformattedLines  bool
}

type Option func(*Options)

// Header is present.
func Header() Option {
	return func(opts *Options) {
		opts.Header = true
	}
}

// Separator of the TSV.
func Separator(separator string) Option {
	return func(opts *Options) {
		opts.Separator = UnescapeSeparator(separator)
	}
}

// Fields on which the TSV can be sorted.
func Fields(fields ...string) Option {
	return func(opts *Options) {
		opts.Fields = fields
	}
}

// DropEmptyIndexedFields removes the lines where the comparable is empty.
func DropEmptyIndexedFields() Option {
	return func(opts *Options) {
		opts.DropEmptyIndexedFields = true
	}
}

// SkipMalformattedLines ignores mal-formatted lines.
func SkipMalformattedLines() Option {
	return func(opts *Options) {
		opts.SkipMalformattedLines = true
	}
}
