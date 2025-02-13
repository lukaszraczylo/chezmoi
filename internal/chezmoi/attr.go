package chezmoi

import (
	"io/fs"
	"strings"

	"github.com/rs/zerolog"
)

// A SourceFileTargetType is a the type of a target represented by a file in the
// source state. A file in the source state can represent a file, script, or
// symlink in the target state.
type SourceFileTargetType int

// Source file types.
const (
	SourceFileTypeCreate SourceFileTargetType = iota
	SourceFileTypeFile
	SourceFileTypeModify
	SourceFileTypeRemove
	SourceFileTypeScript
	SourceFileTypeSymlink
)

var sourceFileTypeStrs = map[SourceFileTargetType]string{
	SourceFileTypeCreate:  "create",
	SourceFileTypeFile:    "file",
	SourceFileTypeModify:  "modify",
	SourceFileTypeRemove:  "remove",
	SourceFileTypeScript:  "script",
	SourceFileTypeSymlink: "symlink",
}

// A ScriptOrder defines when a script should be executed.
type ScriptOrder int

// Script orders.
const (
	ScriptOrderBefore ScriptOrder = -1
	ScriptOrderDuring ScriptOrder = 0
	ScriptOrderAfter  ScriptOrder = 1
)

// A ScriptCondition defines under what conditions a script should be executed.
type ScriptCondition string

// Script conditions.
const (
	ScriptConditionNone     ScriptCondition = ""
	ScriptConditionAlways   ScriptCondition = "always"
	ScriptConditionOnce     ScriptCondition = "once"
	ScriptConditionOnChange ScriptCondition = "onchange"
)

// DirAttr holds attributes parsed from a source directory name.
type DirAttr struct {
	TargetName string
	Exact      bool
	External   bool
	Private    bool
	ReadOnly   bool
	Remove     bool
}

// A FileAttr holds attributes parsed from a source file name.
type FileAttr struct {
	TargetName string
	Condition  ScriptCondition
	Type       SourceFileTargetType
	Order      ScriptOrder
	Empty      bool
	Encrypted  bool
	Executable bool
	Private    bool
	ReadOnly   bool
	Template   bool
}

// parseDirAttr parses a single directory name in the source state.
func parseDirAttr(name string) DirAttr {
	name, remove := strings.CutPrefix(name, removePrefix)
	name, external := strings.CutPrefix(name, externalPrefix)
	name, exact := strings.CutPrefix(name, exactPrefix)
	name, private := strings.CutPrefix(name, privatePrefix)
	name, readOnly := strings.CutPrefix(name, readOnlyPrefix)
	switch {
	case strings.HasPrefix(name, dotPrefix):
		name = "." + name[len(dotPrefix):]
	case strings.HasPrefix(name, literalPrefix):
		name = name[len(literalPrefix):]
	}
	return DirAttr{
		TargetName: name,
		Exact:      exact,
		External:   external,
		Private:    private,
		ReadOnly:   readOnly,
		Remove:     remove,
	}
}

// MarshalZerologObject implements
// github.com/rs/zerolog.ObjectMarshaler.MarshalZerologObject.
func (da DirAttr) MarshalZerologObject(e *zerolog.Event) {
	e.Str("TargetName", da.TargetName)
	e.Bool("Exact", da.Exact)
	e.Bool("External", da.External)
	e.Bool("Private", da.Private)
	e.Bool("ReadOnly", da.ReadOnly)
	e.Bool("Remove", da.Remove)
}

// SourceName returns da's source name.
func (da DirAttr) SourceName() string {
	sourceName := ""
	if da.Remove {
		sourceName += removePrefix
	}
	if da.External {
		sourceName += externalPrefix
	}
	if da.Exact {
		sourceName += exactPrefix
	}
	if da.Private {
		sourceName += privatePrefix
	}
	if da.ReadOnly {
		sourceName += readOnlyPrefix
	}
	switch {
	case strings.HasPrefix(da.TargetName, "."):
		sourceName += dotPrefix + da.TargetName[len("."):]
	case dirPrefixRx.MatchString(da.TargetName):
		sourceName += literalPrefix + da.TargetName
	default:
		sourceName += da.TargetName
	}
	return sourceName
}

// perm returns da's file mode.
func (da DirAttr) perm() fs.FileMode {
	perm := fs.ModePerm
	if da.Private {
		perm &^= 0o77
	}
	if da.ReadOnly {
		perm &^= 0o222
	}
	return perm
}

// parseFileAttr parses a source file name in the source state.
func parseFileAttr(sourceName, encryptedSuffix string) FileAttr {
	var (
		sourceFileType = SourceFileTypeFile
		name           = sourceName
		condition      = ScriptConditionNone
		empty          = false
		encrypted      = false
		executable     = false
		order          = ScriptOrderDuring
		private        = false
		readOnly       = false
		template       = false
	)
	switch {
	case strings.HasPrefix(name, createPrefix):
		sourceFileType = SourceFileTypeCreate
		name = name[len(createPrefix):]
		name, encrypted = strings.CutPrefix(name, encryptedPrefix)
		name, private = strings.CutPrefix(name, privatePrefix)
		name, readOnly = strings.CutPrefix(name, readOnlyPrefix)
		name, empty = strings.CutPrefix(name, emptyPrefix)
		name, executable = strings.CutPrefix(name, executablePrefix)
	case strings.HasPrefix(name, removePrefix):
		sourceFileType = SourceFileTypeRemove
		name = name[len(removePrefix):]
	case strings.HasPrefix(name, runPrefix):
		sourceFileType = SourceFileTypeScript
		name = name[len(runPrefix):]
		switch {
		case strings.HasPrefix(name, oncePrefix):
			name = name[len(oncePrefix):]
			condition = ScriptConditionOnce
		case strings.HasPrefix(name, onChangePrefix):
			name = name[len(onChangePrefix):]
			condition = ScriptConditionOnChange
		default:
			condition = ScriptConditionAlways
		}
		switch {
		case strings.HasPrefix(name, beforePrefix):
			name = name[len(beforePrefix):]
			order = ScriptOrderBefore
		case strings.HasPrefix(name, afterPrefix):
			name = name[len(afterPrefix):]
			order = ScriptOrderAfter
		}
	case strings.HasPrefix(name, symlinkPrefix):
		sourceFileType = SourceFileTypeSymlink
		name = name[len(symlinkPrefix):]
	case strings.HasPrefix(name, modifyPrefix):
		sourceFileType = SourceFileTypeModify
		name = name[len(modifyPrefix):]
		name, encrypted = strings.CutPrefix(name, encryptedPrefix)
		name, private = strings.CutPrefix(name, privatePrefix)
		name, readOnly = strings.CutPrefix(name, readOnlyPrefix)
		name, executable = strings.CutPrefix(name, executablePrefix)
	default:
		name, encrypted = strings.CutPrefix(name, encryptedPrefix)
		name, private = strings.CutPrefix(name, privatePrefix)
		name, readOnly = strings.CutPrefix(name, readOnlyPrefix)
		name, empty = strings.CutPrefix(name, emptyPrefix)
		name, executable = strings.CutPrefix(name, executablePrefix)
	}
	switch {
	case strings.HasPrefix(name, dotPrefix):
		name = "." + name[len(dotPrefix):]
	case strings.HasPrefix(name, literalPrefix):
		name = name[len(literalPrefix):]
	}
	if encrypted {
		name, _ = strings.CutSuffix(name, encryptedSuffix)
	}
	switch {
	case strings.HasSuffix(name, literalSuffix):
		name = name[:len(name)-len(literalSuffix)]
	case strings.HasSuffix(name, TemplateSuffix):
		name = name[:len(name)-len(TemplateSuffix)]
		template = true
		name, _ = strings.CutSuffix(name, literalSuffix)
	}
	return FileAttr{
		TargetName: name,
		Type:       sourceFileType,
		Condition:  condition,
		Empty:      empty,
		Encrypted:  encrypted,
		Executable: executable,
		Order:      order,
		Private:    private,
		ReadOnly:   readOnly,
		Template:   template,
	}
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (fa FileAttr) MarshalZerologObject(e *zerolog.Event) {
	e.Str("TargetName", fa.TargetName)
	e.Str("Type", sourceFileTypeStrs[fa.Type])
	e.Str("Condition", string(fa.Condition))
	e.Bool("Empty", fa.Empty)
	e.Bool("Encrypted", fa.Encrypted)
	e.Bool("Executable", fa.Executable)
	e.Int("Order", int(fa.Order))
	e.Bool("Private", fa.Private)
	e.Bool("ReadOnly", fa.ReadOnly)
	e.Bool("Template", fa.Template)
}

// SourceName returns fa's source name.
func (fa FileAttr) SourceName(encryptedSuffix string) string {
	sourceName := ""
	switch fa.Type {
	case SourceFileTypeCreate:
		sourceName = createPrefix
		if fa.Encrypted {
			sourceName += encryptedPrefix
		}
		if fa.Private {
			sourceName += privatePrefix
		}
		if fa.ReadOnly {
			sourceName += readOnlyPrefix
		}
		if fa.Empty {
			sourceName += emptyPrefix
		}
		if fa.Executable {
			sourceName += executablePrefix
		}
	case SourceFileTypeFile:
		if fa.Encrypted {
			sourceName += encryptedPrefix
		}
		if fa.Private {
			sourceName += privatePrefix
		}
		if fa.ReadOnly {
			sourceName += readOnlyPrefix
		}
		if fa.Empty {
			sourceName += emptyPrefix
		}
		if fa.Executable {
			sourceName += executablePrefix
		}
	case SourceFileTypeModify:
		sourceName = modifyPrefix
		if fa.Encrypted {
			sourceName += encryptedPrefix
		}
		if fa.Private {
			sourceName += privatePrefix
		}
		if fa.ReadOnly {
			sourceName += readOnlyPrefix
		}
		if fa.Executable {
			sourceName += executablePrefix
		}
	case SourceFileTypeRemove:
		sourceName = removePrefix
	case SourceFileTypeScript:
		sourceName = runPrefix
		switch fa.Condition {
		case ScriptConditionOnce:
			sourceName += oncePrefix
		case ScriptConditionOnChange:
			sourceName += onChangePrefix
		}
		switch fa.Order {
		case ScriptOrderBefore:
			sourceName += beforePrefix
		case ScriptOrderAfter:
			sourceName += afterPrefix
		}
	case SourceFileTypeSymlink:
		sourceName = symlinkPrefix
	}
	switch {
	case strings.HasPrefix(fa.TargetName, "."):
		sourceName += dotPrefix + fa.TargetName[len("."):]
	case filePrefixRx.MatchString(fa.TargetName):
		sourceName += literalPrefix + fa.TargetName
	default:
		sourceName += fa.TargetName
	}
	if fileSuffixRx.MatchString(fa.TargetName) {
		sourceName += literalSuffix
	}
	if fa.Template {
		sourceName += TemplateSuffix
	}
	if fa.Encrypted {
		sourceName += encryptedSuffix
	}
	return sourceName
}

// perm returns fa's permissions.
func (fa FileAttr) perm() fs.FileMode {
	perm := fs.FileMode(0o666)
	if fa.Executable {
		perm |= 0o111
	}
	if fa.Private {
		perm &^= 0o77
	}
	if fa.ReadOnly {
		perm &^= 0o222
	}
	return perm
}
