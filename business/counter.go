package business

import (
	"context"
	"database/sql"
	"errors"
	"github.com/angel-one/go-utils/log"
	"github.com/sinhashubham95/go-example-project/constants"
	"github.com/sinhashubham95/go-example-project/utils/configs"
	"github.com/sinhashubham95/go-example-project/utils/database"
	"time"
)

// CreateCounter is used to create a new counter against this key
func CreateCounter(ctx context.Context, key string) error {
	ctx, cancel := getCounterContext(ctx)
	defer cancel()

	db := database.Get()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// check if counter already exists
	_, exists := doesCounterExist(ctx, tx, key)
	if exists {
		return errors.New("counter already exists")
	}

	// now that it does not exist, create a new one
	_, err = tx.ExecContext(ctx, "insert into counter (id, count) values (?, 0)", key)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// IncrementCounter is used to increment the count for the counter if it already exists
func IncrementCounter(ctx context.Context, key string) error {
	ctx, cancel := getCounterContext(ctx)
	defer cancel()

	db := database.Get()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// check if counter already exists
	count, exists := doesCounterExist(ctx, tx, key)
	if !exists {
		return errors.New("counter does not exist")
	}

	// now that counter exists, increment it
	_, err = tx.ExecContext(ctx, "update counter set count = ? where id = ?", count+1, key)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// DecrementCounter is used to decrement the count for the counter if it already exists
func DecrementCounter(ctx context.Context, key string) error {
	ctx, cancel := getCounterContext(ctx)
	defer cancel()

	db := database.Get()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// check if counter already exists
	count, exists := doesCounterExist(ctx, tx, key)
	if !exists {
		return errors.New("counter does not exist")
	}

	// now that counter exists, increment it
	if count > 0 {
		_, err = tx.ExecContext(ctx, "update counter set count = ? where id = ?", count-1, key)
		if err != nil {
			return err
		}
	} else {
		log.Info(ctx).Msgf("count for key %s is 0", key)
	}

	return tx.Commit()
}

// CurrentCount is used to get the current value of counter if it exists
func CurrentCount(ctx context.Context, key string) (int, error) {
	ctx, cancel := getCounterContext(ctx)
	defer cancel()

	db := database.Get()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return 0, err
	}

	// check if counter already exists
	count, exists := doesCounterExist(ctx, tx, key)
	if !exists {
		return 0, errors.New("counter does not exist")
	}

	return count, nil
}

func getCounterContext(ctx context.Context) (context.Context, context.CancelFunc) {
	counterQueryTimeoutInMillis, err := configs.Get().GetInt(constants.CounterConfig,
		constants.CounterQueryTimeoutInMillisKey)
	if err != nil {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, time.Millisecond*time.Duration(counterQueryTimeoutInMillis))
}

func doesCounterExist(ctx context.Context, tx *sql.Tx, key string) (int, bool) {
	row := tx.QueryRowContext(ctx, "select count from counter where id = ?", key)

	if row == nil || errors.Is(row.Err(), sql.ErrNoRows) {
		return 0, false
	}

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, false
	}

	return count, true
}
