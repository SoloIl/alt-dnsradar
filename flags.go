package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func initFlags() {
	flag.Usage = printUsage

	flagSettings.LogName = flag.String(
		"l",
		"",
		msgLogFileFlag(),
	)

	flagSettings.Resolver = flag.String(
		"resolver",
		"https://dns.google/dns-query",
		msgResolverFlag(),
	)

	flagSettings.RequestsTimeout = flag.Int(
		"timeout",
		3,
		msgTimeoutFlag(),
	)

	flagSettings.ShowAll = flag.Bool(
		"all",
		false,
		msgShowAllFlag(),
	)

	flagSettings.Verbose = flag.Bool(
		"v",
		false,
		msgVerboseFlag(),
	)

	flagSettings.QuietMode = flag.Bool(
		"q",
		false,
		msgQuietFlag(),
	)

	flagSettings.NoColor = flag.Bool(
		"no-color",
		false,
		msgNoColorFlag(),
	)

	flagSettings.Lang = flag.String(
		"lang",
		"en",
		msgLangFlag(),
	)

	domain, extras, reorderedArgs := splitArgs(os.Args[1:])
	flag.CommandLine.Parse(reorderedArgs)

	if domain == "" {
		printUsage()
		os.Exit(1)
	}

	if len(extras) > 0 {
		_, _ = fmt.Fprint(os.Stderr, msgUnexpectedArgs(extras))
		printUsage()
		os.Exit(1)
	}

	flagSettings.URL = &domain
}

func splitArgs(args []string) (string, []string, []string) {
	var domain string
	var extras []string
	var out []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") {
			out = append(out, arg)

			if flagConsumesValue(arg) && i+1 < len(args) {
				out = append(out, args[i+1])
				i++
			}
			continue
		}

		if domain == "" {
			domain = arg
		} else {
			extras = append(extras, arg)
		}
	}

	return domain, extras, out
}

func flagConsumesValue(arg string) bool {
	if strings.Contains(arg, "=") {
		return false
	}

	switch arg {
	case "-l", "--l", "-resolver", "--resolver", "-timeout", "--timeout", "-lang", "--lang":
		return true
	default:
		return false
	}
}

func printUsage() {
	binary := thisFilename()

	fmt.Println(msgUsageTitle())
	fmt.Printf("  %s <domain> [flags]\n", binary)
	fmt.Println("")
	fmt.Println(msgExamplesTitle())
	fmt.Println(msgRecommendedExample(binary))
	fmt.Println(msgSecondaryExample(binary))
	fmt.Println("")
	fmt.Println(msgDefaultBehaviorTitle())
	for _, line := range msgDefaultBehaviorLines() {
		fmt.Printf("  - %s\n", line)
	}
	fmt.Println("")

	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("  -%s\n", f.Name)
		desc := f.Usage
		if meta := flagMeta(f.Name, f.DefValue); meta != "" {
			desc = fmt.Sprintf("%s (%s)", desc, meta)
		}
		fmt.Printf("    %s\n", desc)
	})
}

func flagMeta(name string, def string) string {
	switch name {
	case "l":
		return "optional, string"
	case "resolver":
		return fmt.Sprintf("default %q, string", def)
	case "timeout":
		return fmt.Sprintf("default %s, int", def)
	case "lang":
		return fmt.Sprintf("default %q, string", def)
	default:
		return ""
	}
}
