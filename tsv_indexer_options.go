package iosupport

// Options contains information for TSV interations.
type Options struct {
	Header                 bool
	Separator              byte
	Fields                 []string
	DropEmptyIndexedFields bool
	SkipMalformattedLines  bool
	LineThreshold          int
	Swapper                *Swapper
	LazyQuotes             bool
}

// Option is a function used in the Functional Options pattern.
type Option func(*Options)

// Header is present or not.
func Header(header bool) Option {
	return func(opts *Options) {
		opts.Header = header
	}
}

// HasHeader is present.
func HasHeader() Option {
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

// LineThreshold defines the number of file's seekers. One seeker per LineThreshold.
// The number of seekers increase the Transfer speed during sort.
// default: 2500000
func LineThreshold(threshold int) Option {
	return func(opts *Options) {
		opts.LineThreshold = threshold
	}
}

// SwapperOpts defines the memory swapper of the TSV indexer.
// The number of seekers increase the Transfer speed during sort.
func SwapperOpts(limit uint64, basepath string) Option {
	return func(opts *Options) {
		opts.Swapper = NewSwapper(limit, basepath)
	}
}

// LazyQuotesMode allows lazy quotes in CSV.
func LazyQuotesMode() Option {
	return func(opts *Options) {
		opts.LazyQuotes = true
	}
}
