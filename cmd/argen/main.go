package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	argen "github.com/Educentr/go-activerecord/v3/internal/app"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
	"golang.org/x/mod/modfile"
)

// ldflags
var (
	Version     string
	BuildTime   string
	BuildOS     string
	BuildCommit string
)

func getAppInfo() *ds.AppInfo {
	return ds.NewAppInfo().
		WithVersion(Version).
		WithBuildTime(BuildTime).
		WithBuildOS(BuildOS).
		WithBuildCommit(BuildCommit)
}

func main() {
	ctx := context.Background()
	path := flag.String("path", "./repository", "Path to repository dir")
	fixturePath := flag.String("fixture_path", "", "Path to stores of tested fixtures")
	declarationDir := flag.String("declaration", "declaration", "declaration subdir")
	destinationDir := flag.String("destination", "generated", "generation subdir")
	moduleName := flag.String("module", "", "module name from go.mod")
	schemaPath := flag.String("schema_path", "", "Path for DDL schema and migrations (e.g. etc/database)")
	skipMigration := flag.Bool("skip_migration", false, "Skip migration generation")
	version := flag.Bool("version", false, "print version")
	flag.Parse()

	if *version {
		fmt.Printf("Version %s; BuildCommit: %s\n", Version, BuildCommit)
		os.Exit(0)
	}

	srcDir := filepath.Join(*path, *declarationDir)
	dstDir := filepath.Join(*path, *destinationDir)

	if *moduleName == "" {
		goModBytes, err := os.ReadFile("go.mod")
		if err != nil {
			log.Fatalf("error get mod.go")
		}

		*moduleName = modfile.ModulePath(goModBytes)
		if *moduleName == "" {
			log.Fatalf("can't determine module name")
		}
	}

	opts := argen.Options{
		SchemaPath:    *schemaPath,
		SkipMigration: *skipMigration,
	}

	gen, err := argen.Init(ctx, getAppInfo(), srcDir, dstDir, *fixturePath, *moduleName, opts)
	if err != nil {
		log.Fatalf("error initialization: %s", err)
	}

	if err := gen.Run(); err != nil {
		log.Fatalf("error generate repository: %s", err)
	}
}
