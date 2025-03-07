package database

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const SlowSqlDuration = time.Second

var _ logger.Interface = (*Tracer)(nil)

type Tracer struct {
	logger *zap.SugaredLogger
}

func NewTracer(logger *zap.SugaredLogger) *Tracer {
	return &Tracer{logger: logger}
}

func (d *Tracer) LogMode(level logger.LogLevel) logger.Interface {
	return d
}

func (d *Tracer) Info(ctx context.Context, s string, i ...interface{}) {
	d.logger.Infof(s, i)
}

func (d *Tracer) Warn(ctx context.Context, s string, i ...interface{}) {
	d.logger.Warnf(s, i)
}

func (d *Tracer) Error(ctx context.Context, s string, i ...interface{}) {
	d.logger.Errorf(s, i)
}

func (d *Tracer) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil:
		sql, rows := fc()
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			d.logger.Errorf("[%.3fms] [rows:%v] %s err for %s", float64(elapsed.Nanoseconds())/1e6, rows, sql, err)
		}
	case elapsed > SlowSqlDuration:
		sql, rows := fc()
		d.logger.Warnf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	default:
		//sql, rows := fc()
		//d.logger.Infof("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}
