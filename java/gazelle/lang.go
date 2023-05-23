package gazelle

import (
	"os"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/logconfig"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/rs/zerolog"
)

// javaLang is a language.Language implementation for Java.
type javaLang struct {
	config.Configurer
	resolve.Resolver

	parser        *javaparser.Runner
	logger        zerolog.Logger
	javaLogLevel  string
	mavenResolver maven.Resolver

	// javaPackageCache is used for module granularity support
	// Key is the path to the java package from the Bazel workspace root.
	javaPackageCache map[string]*java.Package
}

func NewLanguage() language.Language {
	goLevel, javaLevel := logconfig.LogLevel()

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(goLevel)
	logger.Print("creating java language")

	l := javaLang{
		logger:           logger,
		javaLogLevel:     javaLevel,
		javaPackageCache: make(map[string]*java.Package),
	}

	l.Configurer = NewConfigurer(&l)
	l.Resolver = NewResolver(&l)

	return &l
}

var kindWithRuntimeDeps = rule.KindInfo{
	NonEmptyAttrs: map[string]bool{
		"deps": true,
		"srcs": true,
	},
	MergeableAttrs: map[string]bool{"srcs": true},
	ResolveAttrs: map[string]bool{
		"deps":         true,
		"runtime_deps": true,
	},
}
var kindWithoutRuntimeDeps = rule.KindInfo{
	NonEmptyAttrs: map[string]bool{
		"deps": true,
		"srcs": true,
	},
	MergeableAttrs: map[string]bool{"srcs": true},
	ResolveAttrs: map[string]bool{
		"deps": true,
	},
}

var javaLibraryKind = rule.KindInfo{
	NonEmptyAttrs: map[string]bool{
		"deps":    true,
		"exports": true,
		"srcs":    true,
	},
	MergeableAttrs: map[string]bool{"srcs": true},
	ResolveAttrs: map[string]bool{
		"deps":         true,
		"exports":      true,
		"runtime_deps": true,
	},
}

func (l javaLang) Kinds() map[string]rule.KindInfo {
	return map[string]rule.KindInfo{
		"java_binary":        kindWithRuntimeDeps,
		"java_junit5_test":   kindWithRuntimeDeps,
		"java_library":       javaLibraryKind,
		"java_test":          kindWithRuntimeDeps,
		"java_test_suite":    kindWithRuntimeDeps,
		"java_proto_library": kindWithoutRuntimeDeps,
		"java_grpc_library":  kindWithoutRuntimeDeps,
		"unit_package": 	  kindWithRuntimeDeps,
		"unit_test": 	  	  kindWithRuntimeDeps,
		"int_test": 		  kindWithRuntimeDeps,
		"test_test":		  kindWithRuntimeDeps,
		"library_package": 	  javaLibraryKind,
		"test_package": 	  kindWithRuntimeDeps,	  
	}
}

func isTestRule(kind string) bool {
	var test_rule_found bool = false
	test_kinds := [...]string{
		"java_junit5_test", 
		"java_test", 
		"java_test_suite", 
		"unit_package", 
		"unit_test", 
		"int_test",
		"test_test",
		"test_package",
	}
	for i :=0; i < len(test_kinds); i++ {
		if test_kinds[i] == string {
			test_rule_found = true
			break
		}
	}
	return test_rule_found
}

var javaLoads = []rule.LoadInfo{
	{
		Name: "@io_grpc_grpc_java//:java_grpc_library.bzl",
		Symbols: []string{
			"java_grpc_library",
		},
	},
	{
		Name: "@rules_java//java:defs.bzl",
		Symbols: []string{
			"java_binary",
			"java_library",
			"java_proto_library",
			"java_test",
		},
	},
	{
		Name: "@contrib_rules_jvm//java:defs.bzl",
		Symbols: []string{
			"java_junit5_test",
			"java_test_suite",
		},
	},
	{
		Name: "@10gen_mms//server/src/unit:rules.bzl",
		Symbols: []string{
			"unit_test",
			"unit_package",
			"library_package",
		}
	},
	{
		Name: "@10gen_mms//server/src/test:rules.bzl",
		Symbols: []string{
			"int_test",
			"test_test",
			"test_package",
		}
	}
}

func (l javaLang) Loads() []rule.LoadInfo {
	return javaLoads
}

func (l javaLang) Fix(c *config.Config, f *rule.File) {}

func (l javaLang) DoneGeneratingRules() {
	l.parser.ServerManager().Shutdown()
}
