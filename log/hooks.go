package log

import (
	"github.com/sirupsen/logrus"
)

// ContextValueFieldHook add a FieldKey=ContextValue(ContextKey) field
// (if exists) to the log entry
type ContextValueFieldHook struct {
	FieldKey   string // default: "context"
	ContextKey string
}

func (c ContextValueFieldHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (c ContextValueFieldHook) Fire(entry *logrus.Entry) error {
	//fmt.Printf("[DEBUG] ContextValueFieldHook: hook=%#v, context=%#v\n", c, entry.Context)

	if c.ContextKey == "" { // nothing to do
		return nil
	}
	if c.FieldKey == "" {
		c.FieldKey = "context"
	}
	if entry.Context == nil {
		return nil
	}
	value := entry.Context.Value(c.ContextKey)
	if value == nil {
		return nil
	}

	// cannot WithField in hook
	//entry = entry.WithField(c.FieldKey, entry.Context.Value(c.ContextKey))
	entry.Data[c.FieldKey] = entry.Context.Value(c.ContextKey)
	return nil
}

// RequestIDHook add a context="request_id" field to the log entry
func RequestIDHook() logrus.Hook {
	return ContextValueFieldHook{
		ContextKey: "request_id",
	}
}
