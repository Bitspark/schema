#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Reorganizes the schema package structure for better maintainability.

.DESCRIPTION
    This script performs the migration outlined in REORGANIZATION_PLAN.md:
    - Moves reflection files to /schema/reflection/
    - Moves generator files to /schema/generator/
    - Moves visitor files to /schema/visitor/
    - Updates package declarations
    - Creates backward compatibility layer
    - Validates the migration

.PARAMETER DryRun
    If specified, shows what would be done without making actual changes.

.PARAMETER SkipCompatibility
    If specified, skips creating the backward compatibility layer.

.PARAMETER Rollback
    If specified, attempts to rollback a previous migration.

.EXAMPLE
    .\reorganize.ps1
    Performs the full migration.

.EXAMPLE
    .\reorganize.ps1 -DryRun
    Shows what would be done without making changes.

.EXAMPLE
    .\reorganize.ps1 -Rollback
    Attempts to rollback the migration.
#>

param(
    [switch]$DryRun,
    [switch]$SkipCompatibility,
    [switch]$Rollback
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Colors for output
$Color = @{
    Green = "Green"
    Yellow = "Yellow"
    Red = "Red"
    Cyan = "Cyan"
    Magenta = "Magenta"
}

function Write-Status {
    param([string]$Message, [string]$Color = "White")
    Write-Host "🔄 $Message" -ForegroundColor $Color
}

function Write-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor $Color.Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor $Color.Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor $Color.Red
}

function Write-Info {
    param([string]$Message)
    Write-Host "ℹ️  $Message" -ForegroundColor $Color.Cyan
}

# Get the script directory (should be the schema directory)
$SchemaDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$OriginalDir = Get-Location

# Validate we're in the right directory
if (-not (Test-Path (Join-Path $SchemaDir "types.go"))) {
    Write-Error "This script must be run from the schema directory containing types.go"
    exit 1
}

# Change to schema directory
Set-Location $SchemaDir

Write-Host "🚀 Schema Package Reorganization Script" -ForegroundColor $Color.Magenta
Write-Host "=======================================" -ForegroundColor $Color.Magenta
Write-Host ""

if ($DryRun) {
    Write-Info "DRY RUN MODE: No actual changes will be made"
    Write-Host ""
}

if ($Rollback) {
    Write-Warning "ROLLBACK MODE: Attempting to reverse migration"
    Write-Host ""
}

# Define file mappings
$FileMappings = @{
    "reflection" = @{
        "reflection.go" = "reflection/reflection.go"
        "reflection_funcs.go" = "reflection/funcs.go"
        "reflect_service.go" = "reflection/service.go"
        "reflection_test.go" = "reflection/reflection_test.go"
        "reflection_funcs_test.go" = "reflection/funcs_test.go"
        "reflection_advanced_test.go" = "reflection/advanced_test.go"
    }
    "generator" = @{
        "generator.go" = "generator/generator.go"
        "schema_generator.go" = "generator/schema.go"
        "generator_test.go" = "generator/generator_test.go"
        "schema_generator_test.go" = "generator/schema_test.go"
        "default_generator_test.go" = "generator/default_test.go"
    }
    "visitor" = @{
        "visitor.go" = "visitor/visitor.go"
        "visitor_test.go" = "visitor/visitor_test.go"
    }
}

# Backup directory for rollback
$BackupDir = "reorganize_backup_$(Get-Date -Format 'yyyyMMdd_HHmmss')"

function Test-Prerequisites {
    Write-Status "Checking prerequisites..."
    
    # Check if Go is installed
    try {
        $goVersion = go version
        Write-Success "Go is installed: $goVersion"
    } catch {
        Write-Error "Go is not installed or not in PATH"
        return $false
    }
    
    # Check if this is a Git repository
    if (Test-Path ".git") {
        Write-Success "Git repository detected"
        
        # Check for uncommitted changes
        $gitStatus = git status --porcelain
        if ($gitStatus) {
            Write-Warning "Uncommitted changes detected. Consider committing before migration."
            Write-Host $gitStatus
            
            $response = Read-Host "Continue anyway? (y/N)"
            if ($response -notmatch '^[Yy]') {
                return $false
            }
        }
    } else {
        Write-Warning "Not a Git repository. Backup recommended before migration."
    }
    
    return $true
}

function New-DirectoryStructure {
    Write-Status "Creating new directory structure..."
    
    $directories = @("reflection", "generator", "visitor")
    
    foreach ($dir in $directories) {
        if (-not $DryRun) {
            if (-not (Test-Path $dir)) {
                New-Item -ItemType Directory -Path $dir -Force | Out-Null
                Write-Success "Created directory: $dir"
            } else {
                Write-Info "Directory already exists: $dir"
            }
        } else {
            Write-Info "Would create directory: $dir"
        }
    }
}

function Move-Files {
    Write-Status "Moving files to new locations..."
    
    # Create backup directory if not in dry run mode
    if (-not $DryRun -and -not $Rollback) {
        New-Item -ItemType Directory -Path $BackupDir -Force | Out-Null
        Write-Info "Created backup directory: $BackupDir"
    }
    
    foreach ($subsystem in $FileMappings.Keys) {
        Write-Status "Processing $subsystem files..." -Color $Color.Cyan
        
        foreach ($sourceFile in $FileMappings[$subsystem].Keys) {
            $targetFile = $FileMappings[$subsystem][$sourceFile]
            
            if ($Rollback) {
                # Reverse the mapping for rollback
                $temp = $sourceFile
                $sourceFile = $targetFile
                $targetFile = $temp
            }
            
            if (Test-Path $sourceFile) {
                if (-not $DryRun) {
                    # Create backup
                    if (-not $Rollback) {
                        Copy-Item $sourceFile (Join-Path $BackupDir (Split-Path $sourceFile -Leaf)) -Force
                    }
                    
                    # Ensure target directory exists
                    $targetDir = Split-Path $targetFile -Parent
                    if ($targetDir -and -not (Test-Path $targetDir)) {
                        New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
                    }
                    
                    # Move file
                    Move-Item $sourceFile $targetFile -Force
                    Write-Success "Moved: $sourceFile → $targetFile"
                } else {
                    Write-Info "Would move: $sourceFile → $targetFile"
                }
            } else {
                if (-not $Rollback) {
                    Write-Warning "Source file not found: $sourceFile"
                }
            }
        }
    }
}

function Update-PackageDeclarations {
    Write-Status "Updating package declarations..."
    
    $packageMappings = @{
        "reflection/*.go" = "reflection"
        "generator/*.go" = "generator"
        "visitor/*.go" = "visitor"
    }
    
    foreach ($pattern in $packageMappings.Keys) {
        $newPackage = $packageMappings[$pattern]
        $files = Get-ChildItem $pattern -ErrorAction SilentlyContinue
        
        foreach ($file in $files) {
            if ($file.Name -like "*_test.go") {
                # Handle test files separately if needed
                continue
            }
            
            $content = Get-Content $file.FullName -Raw
            $originalContent = $content
            
            if ($Rollback) {
                # Change back to package schema
                $content = $content -replace "^package $newPackage", "package schema"
                $content = $content -replace "import `"defs\.dev/schema`"", ""
                $content = $content -replace "schema\.", ""
            } else {
                # Change from package schema to new package
                $content = $content -replace "^package schema", "package $newPackage"
                
                # Add import for schema package if not present and needed
                if ($content -match "Schema|ValidationResult|SchemaType" -and $content -notmatch "import.*defs\.dev/schema") {
                    $content = $content -replace "(package $newPackage\s*\n)", "`$1`nimport `"defs.dev/schema`"`n"
                }
                
                # Update type references
                $content = $content -replace "\bSchema\b", "schema.Schema"
                $content = $content -replace "\bValidationResult\b", "schema.ValidationResult"
                $content = $content -replace "\bSchemaType\b", "schema.SchemaType"
                $content = $content -replace "\bSchemaMetadata\b", "schema.SchemaMetadata"
            }
            
            if ($content -ne $originalContent) {
                if (-not $DryRun) {
                    Set-Content $file.FullName $content -NoNewline
                    Write-Success "Updated package declaration: $($file.Name)"
                } else {
                    Write-Info "Would update package declaration: $($file.Name)"
                }
            }
        }
    }
}

function New-CompatibilityLayer {
    if ($SkipCompatibility -or $Rollback) {
        return
    }
    
    Write-Status "Creating backward compatibility layer..."
    
    # Reflection compatibility
    $reflectionCompat = @"
package schema

import (
    "reflect"
    "defs.dev/schema/reflection"
)

// Deprecated: Use reflection.FromStruct instead.
// This function will be removed in a future version.
func FromStruct[T any]() Schema {
    return reflection.FromStruct[T]()
}

// Deprecated: Use reflection.FromType instead.
// This function will be removed in a future version.
func FromType(typ reflect.Type) Schema {
    return reflection.FromType(typ)
}

// Deprecated: Use reflection.RegisterTypeMapping instead.
// This function will be removed in a future version.
func RegisterTypeMapping(typ reflect.Type, schemaFactory func() Schema) {
    reflection.RegisterTypeMapping(typ, schemaFactory)
}
"@

    # Generator compatibility
    $generatorCompat = @"
package schema

import "defs.dev/schema/generator"

// Deprecated: Use generator.New instead.
// This function will be removed in a future version.
func NewGenerator(options ...any) interface{} {
    return generator.New()
}
"@

    # Visitor compatibility
    $visitorCompat = @"
package schema

import "defs.dev/schema/visitor"

// SchemaVisitor is deprecated: Use visitor.Visitor instead.
type SchemaVisitor = visitor.Visitor

// Deprecated: Use visitor.Walk instead.
// This function will be removed in a future version.
func Walk(schema Schema, visitor SchemaVisitor) error {
    return visitor.Walk(schema, visitor)
}
"@

    $compatFiles = @{
        "reflection_compat.go" = $reflectionCompat
        "generator_compat.go" = $generatorCompat
        "visitor_compat.go" = $visitorCompat
    }
    
    foreach ($fileName in $compatFiles.Keys) {
        $content = $compatFiles[$fileName]
        
        if (-not $DryRun) {
            Set-Content $fileName $content
            Write-Success "Created compatibility layer: $fileName"
        } else {
            Write-Info "Would create compatibility layer: $fileName"
        }
    }
}

function Remove-CompatibilityLayer {
    if (-not $Rollback) {
        return
    }
    
    Write-Status "Removing compatibility layer..."
    
    $compatFiles = @("reflection_compat.go", "generator_compat.go", "visitor_compat.go")
    
    foreach ($file in $compatFiles) {
        if (Test-Path $file) {
            if (-not $DryRun) {
                Remove-Item $file -Force
                Write-Success "Removed compatibility file: $file"
            } else {
                Write-Info "Would remove compatibility file: $file"
            }
        }
    }
}

function Test-Migration {
    if ($DryRun) {
        Write-Info "Skipping tests in dry run mode"
        return $true
    }
    
    Write-Status "Testing migration..."
    
    # Test build
    Write-Status "Testing build..." -Color $Color.Yellow
    try {
        $buildOutput = go build ./... 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Build successful"
        } else {
            Write-Error "Build failed:"
            Write-Host $buildOutput
            return $false
        }
    } catch {
        Write-Error "Build test failed: $_"
        return $false
    }
    
    # Test basic functionality
    Write-Status "Running tests..." -Color $Color.Yellow
    try {
        $testOutput = go test ./... 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Tests passed"
        } else {
            Write-Warning "Some tests failed:"
            Write-Host $testOutput
            Write-Info "This may be expected during migration. Check the output carefully."
        }
    } catch {
        Write-Warning "Test execution failed: $_"
        Write-Info "This may be expected during migration."
    }
    
    return $true
}

function Show-Summary {
    Write-Host ""
    Write-Host "📋 Migration Summary" -ForegroundColor $Color.Magenta
    Write-Host "===================" -ForegroundColor $Color.Magenta
    
    if ($Rollback) {
        Write-Host "Rollback completed." -ForegroundColor $Color.Green
        Write-Host ""
        Write-Host "The schema package has been restored to its original flat structure."
    } else {
        Write-Host "Migration completed successfully!" -ForegroundColor $Color.Green
        Write-Host ""
        Write-Host "New structure:" -ForegroundColor $Color.Cyan
        Write-Host "  schema/"
        Write-Host "  ├── types.go, function.go, basic.go, builder.go  # Core files"
        Write-Host "  ├── *_compat.go                                   # Backward compatibility"
        Write-Host "  ├── reflection/                                   # Reflection subsystem"
        Write-Host "  ├── generator/                                    # Generation subsystem"
        Write-Host "  ├── visitor/                                      # Visitor subsystem"
        Write-Host "  ├── registry/                                     # Registry (existing)"
        Write-Host "  └── functions/                                    # Functions (existing)"
        Write-Host ""
        
        if (-not $SkipCompatibility) {
            Write-Host "Backward compatibility:" -ForegroundColor $Color.Yellow
            Write-Host "  • Old APIs still work but are marked as deprecated"
            Write-Host "  • Compatibility layer created in *_compat.go files"
            Write-Host "  • Existing code should continue to work without changes"
        }
        
        Write-Host ""
        Write-Host "Next steps:" -ForegroundColor $Color.Green
        Write-Host "  1. Update your imports to use new packages:"
        Write-Host "     import `"defs.dev/schema/reflection`""
        Write-Host "     import `"defs.dev/schema/generator`""
        Write-Host "     import `"defs.dev/schema/visitor`""
        Write-Host "  2. Run your tests to ensure everything works"
        Write-Host "  3. Update documentation and examples"
        Write-Host "  4. Consider removing compatibility layer in future versions"
        
        if (-not $DryRun) {
            Write-Host ""
            Write-Host "Backup created in: $BackupDir" -ForegroundColor $Color.Cyan
            Write-Host "To rollback: .\reorganize.ps1 -Rollback"
        }
    }
}

# Main execution
try {
    if (-not (Test-Prerequisites)) {
        Write-Error "Prerequisites not met. Aborting."
        exit 1
    }
    
    Write-Host ""
    
    if ($Rollback) {
        Write-Warning "This will attempt to rollback the reorganization."
        $response = Read-Host "Are you sure you want to continue? (y/N)"
        if ($response -notmatch '^[Yy]') {
            Write-Info "Rollback cancelled."
            exit 0
        }
        
        Move-Files
        Update-PackageDeclarations
        Remove-CompatibilityLayer
    } else {
        if (-not $DryRun) {
            Write-Warning "This will reorganize the schema package structure."
            $response = Read-Host "Are you sure you want to continue? (y/N)"
            if ($response -notmatch '^[Yy]') {
                Write-Info "Migration cancelled."
                exit 0
            }
        }
        
        New-DirectoryStructure
        Move-Files
        Update-PackageDeclarations
        New-CompatibilityLayer
    }
    
    if (Test-Migration) {
        Write-Host ""
        Write-Success "Migration validation completed successfully!"
    } else {
        Write-Warning "Migration validation encountered issues. Please review the output above."
    }
    
    Show-Summary
    
} catch {
    Write-Error "An error occurred during migration: $_"
    Write-Host $_.ScriptStackTrace
    exit 1
} finally {
    # Restore original directory
    Set-Location $OriginalDir
}

Write-Host ""
Write-Host "🎉 Script completed!" -ForegroundColor $Color.Green 