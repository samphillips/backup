package progress

import (
	"github.com/cheggaaa/pb"
)

var (
	tmpl        = `{{ green "[INFO]" }} {{ bar . "[" "-" (cycle . "↖" "↗" "↘" "↙" ) "." "]"}} {{percent .}} {{etime .}}`
	progressBar = pb.ProgressBarTemplate(tmpl)
)

// Start starts a progress bar and return a pointer to it
func Start(limit int) *pb.ProgressBar {
	return progressBar.Start(limit)
}
