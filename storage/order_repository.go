package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/abhilashdk2016/transactional-outbox-pattern/models"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
)

type Repository struct {
	DB *pgx.Conn
}

func (r *Repository) CreateOrder(ctx *fiber.Ctx) error {
	var id uint
	order := models.Orders{}
	err := ctx.BodyParser(&order)
	if err != nil {
		ctx.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{"message": "request failed"})
		return err
	}
	tx, _ := r.DB.BeginTx(context.Background(), pgx.TxOptions{})
	defer tx.Rollback(ctx.Context())
	_ = tx.QueryRow(context.Background(), "INSERT INTO orders(customer_id, quantity, price) VALUES($1, $2, $3) RETURNING id", order.CustomerId, order.Quantity, order.Price).Scan(&id)
	order.ID = id
	fmt.Println(order)
	payload, _ := json.Marshal(&order)
	outbox := models.Outbox{
		Payload:     string(payload),
		IsProcessed: false,
	}
	fmt.Println(outbox)
	_, execErr := tx.Exec(context.Background(), "INSERT INTO outbox(payload, is_processed) VALUES($1, $2)", outbox.Payload, outbox.IsProcessed)
	if execErr != nil {
		fmt.Println(execErr.Error())
	}
	if err := tx.Commit(context.Background()); err != nil {
		fmt.Println(err.Error())
		return fmt.Errorf("tx.Commit %w", err)
	}

	ctx.Status(http.StatusOK).JSON(&fiber.Map{"message": "Order has been created"})
	return nil
}

func (r *Repository) GetOrders(ctx *fiber.Ctx) error {
	orders := []models.Orders{}
	rows, err := r.DB.Query(context.Background(), "SELECT id, customer_id, quantity, price FROM orders")
	if err != nil {
		fmt.Println("row.Scan", err)
		return nil
	}
	defer rows.Close()

	if rows.Err() != nil {
		fmt.Println("row.Err()", err)
		return nil
	}

	for rows.Next() {
		order := models.Orders{}
		err := rows.Scan(&order.ID, &order.CustomerId, &order.Quantity, &order.Price)
		if err != nil {
			return fmt.Errorf("unable to scan row: %w", err)
		}
		orders = append(orders, order)
	}
	ctx.Status(http.StatusOK).JSON(&fiber.Map{"message": "Orders fetched", "data": orders})
	return nil
}

func (r *Repository) DeleteOrder(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "Id cannot be empty"})
		return nil
	}
	res, err := r.DB.Exec(context.Background(), "DELETE FROM orders WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("unable to delete order: %w", err)
	}
	ctx.Status(http.StatusOK).JSON(&fiber.Map{"message": "Deleted order", "data": rowsAffected})
	return nil
}

func (r *Repository) GetOrderById(ctx *fiber.Ctx) error {
	order := &models.Orders{}
	id := ctx.Params("id")
	if id == "" {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "Id cannot be empty"})
		return nil
	}

	row := r.DB.QueryRow(context.Background(), `SELECT * FROM orders WHERE Id=$1`, id)
	if err := row.Scan(&order.ID, &order.CustomerId, &order.Quantity, &order.Price); err != nil {
		return fmt.Errorf("no rows in the db: %w", err)
	}
	ctx.Status(http.StatusOK).JSON(&fiber.Map{"message": "Fetched order", "data": order})
	return nil
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/orders", r.CreateOrder)
	api.Delete("orders/:id", r.DeleteOrder)
	api.Get("/orders/:id", r.GetOrderById)
	api.Get("/orders", r.GetOrders)
}
