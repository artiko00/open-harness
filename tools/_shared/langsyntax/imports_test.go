package langsyntax

import (
	"strings"
	"testing"
)

// stripped aplica el pipeline real: StripComments primero, StripImports después.
func stripped(src, ext string) string {
	return StripImports(StripComments(src, ext), ext)
}

func TestStripImports_unknownExtensionPassesThrough(t *testing.T) {
	src := "import foo\nbody"
	if out := StripImports(src, ".txt"); out != src {
		t.Fatalf("extensión desconocida debe pasar intacta, got %q", out)
	}
}

func TestStripImports_singleLineByLanguage(t *testing.T) {
	cases := []struct {
		name, ext, line, marker string
	}{
		{"js import", ".ts", "import { UserService } from './user.service';", "UserService"},
		{"js export from", ".js", "export { Foo } from './foo';", "Foo"},
		{"js require const", ".cjs", "const path = require('path');", "path"},
		{"js require let", ".js", "let fs = require('fs');", "fs"},
		{"js require var", ".jsx", "var os = require('os');", "os"},
		{"py import", ".py", "import collections", "collections"},
		{"py from import", ".py", "from django.db import models", "models"},
		{"go package", ".go", "package main", "main"},
		{"go import", ".go", "import \"fmt\"", "import"},
		{"rb require", ".rb", "require 'json'", "require"},
		{"rb require_relative", ".rb", "require_relative 'helper'", "require_relative"},
		{"rust use", ".rs", "use std::collections::HashMap;", "HashMap"},
		{"rust extern", ".rs", "extern crate serde;", "serde"},
		{"jvm import", ".java", "import java.util.List;", "List"},
		{"jvm package", ".kt", "package com.acme.app", "acme"},
		{"scala import", ".scala", "import scala.collection.mutable", "mutable"},
		{"kts import", ".kts", "import org.gradle.api.Project", "Project"},
		{"php use", ".php", "use App\\Service\\Mailer;", "Mailer"},
		{"php namespace", ".php", "namespace App\\Http;", "Http"},
		{"php require_once", ".php", "require_once 'bootstrap.php';", "require_once"},
		{"c include", ".c", "#include <stdio.h>", "stdio"},
		{"cpp include", ".cpp", "#include \"local.h\"", "include"},
		{"cpp using namespace", ".cc", "using namespace std;", "std"},
		{"header include", ".h", "#include <vector>", "vector"},
		{"hpp include", ".hpp", "#include <memory>", "memory"},
		{"objc import", ".m", "#import <Foundation/Foundation.h>", "Foundation"},
		{"objcpp import", ".mm", "#import <UIKit/UIKit.h>", "UIKit"},
		{"cs using", ".cs", "using System.Collections.Generic;", "Generic"},
		{"cs global using", ".cs", "global using System.Text;", "Text"},
		{"dart import", ".dart", "import 'package:flutter/material.dart';", "import"},
		{"dart export", ".dart", "export 'src/widget.dart';", "export"},
		{"dart part", ".dart", "part 'model.g.dart';", "part"},
		{"dart library", ".dart", "library my_lib;", "my_lib"},
		{"swift import", ".swift", "import Foundation", "Foundation"},
		{"tsx import", ".tsx", "import React from 'react';", "React"},
		{"mjs import", ".mjs", "import x from './x.mjs';", "import"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := stripped(tc.line+"\nkeep_me()", tc.ext)
			if strings.Contains(out, tc.marker) {
				t.Fatalf("%q debía descartarse, quedó %q", tc.line, out)
			}
			if !strings.Contains(out, "keep_me") {
				t.Fatalf("el cuerpo debía conservarse, got %q", out)
			}
		})
	}
}

func TestStripImports_multilineJS(t *testing.T) {
	src := "import {\n  Alpha,\n  Beta,\n} from './mod';\nreal_code()"
	out := stripped(src, ".ts")
	if strings.Contains(out, "Alpha") || strings.Contains(out, "Beta") {
		t.Fatalf("import multilínea debía descartarse completo, got %q", out)
	}
	if !strings.Contains(out, "real_code") {
		t.Fatalf("el cuerpo debía conservarse, got %q", out)
	}
}

func TestStripImports_multilineGoBlock(t *testing.T) {
	src := "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {}"
	out := stripped(src, ".go")
	if strings.Contains(out, "import") {
		t.Fatalf("bloque import debía descartarse, got %q", out)
	}
	if !strings.Contains(out, "func main") {
		t.Fatalf("el cuerpo debía conservarse, got %q", out)
	}
}

func TestStripImports_multilinePythonFromImport(t *testing.T) {
	src := "from app.models import (\n    User,\n    Order,\n)\n\ndef handler():\n    pass"
	out := stripped(src, ".py")
	if strings.Contains(out, "User") || strings.Contains(out, "Order") {
		t.Fatalf("from-import multilínea debía descartarse, got %q", out)
	}
	if !strings.Contains(out, "def handler") {
		t.Fatalf("el cuerpo debía conservarse, got %q", out)
	}
}

func TestStripImports_unbalancedClosersDoNotGoNegative(t *testing.T) {
	src := "import {\n}}\nafter_block()"
	out := stripped(src, ".ts")
	if !strings.Contains(out, "after_block") {
		t.Fatalf("el desbalance no debe tragarse el resto del archivo, got %q", out)
	}
}

func TestStripImports_falsePositives(t *testing.T) {
	cases := []struct {
		name, ext, line, marker string
	}{
		{"cs using resource", ".cs", "using (var stream = File.OpenRead(path))", "stream"},
		{"cs using declaration", ".cs", "using var stream = File.OpenRead(path);", "stream"},
		{"js dynamic import", ".js", "import('./lazy').then(register);", "register"},
		{"identifier prefix", ".ts", "importedValue = compute();", "importedValue"},
		{"use in php closure body", ".php", "usedTotal = 1;", "usedTotal"},
		{"package as identifier", ".go", "packageName := resolve()", "packageName"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := stripped(tc.line, tc.ext)
			if !strings.Contains(out, tc.marker) {
				t.Fatalf("%q no es un import y debía conservarse, got %q", tc.line, out)
			}
		})
	}
}

func TestStripImports_preservesLineCount(t *testing.T) {
	src := "import a from 'a';\nimport {\n  B,\n} from 'b';\n\nfunc()\n"
	out := StripImports(src, ".ts")
	if got, want := strings.Count(out, "\n"), strings.Count(src, "\n"); got != want {
		t.Fatalf("saltos de línea: got %d, want %d", got, want)
	}
}

func TestStripImports_indentedImportIsRecognized(t *testing.T) {
	out := stripped("    import lazy_mod\nkeep()", ".py")
	if strings.Contains(out, "lazy_mod") {
		t.Fatalf("import indentado debía descartarse, got %q", out)
	}
}

func TestStripImports_bareWordAtEndOfLine(t *testing.T) {
	out := StripImports("import\nkeep()", ".go")
	if strings.Contains(out, "import") {
		t.Fatalf("la palabra sola debía descartarse, got %q", out)
	}
}
