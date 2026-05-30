// Command go-lint runs all code-bloat analyzers as a multichecker.
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/AndreyMashukov/go-lint/analyzer/nodbmutintest"
	"github.com/AndreyMashukov/go-lint/analyzer/nodeadguard"
	"github.com/AndreyMashukov/go-lint/analyzer/noenvbranch"
	"github.com/AndreyMashukov/go-lint/analyzer/noerrorwrapbanality"
	"github.com/AndreyMashukov/go-lint/analyzer/nogetenv"
	"github.com/AndreyMashukov/go-lint/analyzer/noinlinecomment"
	"github.com/AndreyMashukov/go-lint/analyzer/nolintdirective"
	"github.com/AndreyMashukov/go-lint/analyzer/nopanicsrc"
	"github.com/AndreyMashukov/go-lint/analyzer/noredundantif"
	"github.com/AndreyMashukov/go-lint/analyzer/norobotgodoc"
	"github.com/AndreyMashukov/go-lint/analyzer/nosilentfallback"
	"github.com/AndreyMashukov/go-lint/analyzer/notimenow"
	"github.com/AndreyMashukov/go-lint/analyzer/notodo"
	"github.com/AndreyMashukov/go-lint/analyzer/notypeonlyassert"
)

func main() {
	multichecker.Main(
		noinlinecomment.Analyzer,
		nolintdirective.Analyzer,
		nogetenv.Analyzer,
		noenvbranch.Analyzer,
		nopanicsrc.Analyzer,
		notimenow.Analyzer,
		notypeonlyassert.Analyzer,
		nodbmutintest.Analyzer,
		norobotgodoc.Analyzer,
		noerrorwrapbanality.Analyzer,
		notodo.Analyzer,
		noredundantif.Analyzer,
		nodeadguard.Analyzer,
		nosilentfallback.Analyzer,
	)
}
