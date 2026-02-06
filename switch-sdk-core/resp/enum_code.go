package resp

type Code int64

type CheckCode func(Code) bool
