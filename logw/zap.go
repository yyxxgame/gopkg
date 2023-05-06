//@File     zap.go
//@Time     2023/05/06
//@Author   #Suyghur,

package logw

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

const callerSkipOffset = 3

type ZapWriter struct {
	logger *zap.Logger
}

func NewZapWriter(opts ...zap.Option) (logx.Writer, error) {
	opts = append(opts, zap.AddCallerSkip(callerSkipOffset))
	logger, err := zap.NewProduction(opts...)
	if err != nil {
		return nil, err
	}

	return &ZapWriter{
		logger: logger,
	}, nil
}

func (w *ZapWriter) Alert(v any) {
	w.logger.Error(fmt.Sprint(v))
}

func (w *ZapWriter) Close() error {
	return w.logger.Sync()
}

func (w *ZapWriter) Debug(v any, fields ...logx.LogField) {
	w.logger.Debug(fmt.Sprint(v), toZapFields(fields...)...)
}

func (w *ZapWriter) Error(v any, fields ...logx.LogField) {
	w.logger.Error(fmt.Sprint(v), toZapFields(fields...)...)
}

func (w *ZapWriter) Info(v any, fields ...logx.LogField) {
	w.logger.Info(fmt.Sprint(v), toZapFields(fields...)...)
}

func (w *ZapWriter) Severe(v any) {
	w.logger.Fatal(fmt.Sprint(v))
}

func (w *ZapWriter) Slow(v any, fields ...logx.LogField) {
	w.logger.Warn(fmt.Sprint(v), toZapFields(fields...)...)
}

func (w *ZapWriter) Stack(v any) {
	w.logger.Error(fmt.Sprint(v), zap.Stack("stack"))
}

func (w *ZapWriter) Stat(v any, fields ...logx.LogField) {
	w.logger.Info(fmt.Sprint(v), toZapFields(fields...)...)
}

func toZapFields(fields ...logx.LogField) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}
	return zapFields
}
