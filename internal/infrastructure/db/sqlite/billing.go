package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/billing"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
)

type BillingRepository struct {
	db *sql.DB
}

func NewBillingRepository(db *sql.DB) *BillingRepository {
	return &BillingRepository{
		db: db,
	}
}

func (r *BillingRepository) executor(ctx context.Context) executor {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (r *BillingRepository) Add(ctx context.Context, i *billing.Invoice) error {
	ex := r.executor(ctx)

	_, err := ex.ExecContext(
		ctx,
		`INSERT INTO invoices(
            id,
            project_id,
            is_invoice,
            status,
            due_at,
            currency_code,
            note,
            total_tax,
            total_discount,
            sub_total,
            version,
            created_at,
            updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		i.ID().String(),
		i.ProjectID().String(),
		i.IsInvoice(),
		i.Status().String(),
		i.DueAt(),
		i.Currency().Code().String(),
		i.Note(),
		i.TotalTax().String(),
		i.TotalDiscount().String(),
		i.SubTotal().String(),
		i.Version(),
		i.CreatedAt(),
		i.UpdatedAt(),
	)
	if err != nil {
		if sqlite3Err, ok := errors.AsType[sqlite3.Error](err); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return fmt.Errorf(
					"billing.BillingRepository.Add: duplicate invoice id: %w",
					err,
				)
			}
		}

		return fmt.Errorf(
			"billing.BillingRepository.Add: %w",
			err,
		)
	}

	items := i.Items()

	if len(items) == 0 {
		return nil
	}

	var sb strings.Builder

	args := make([]any, 0, len(items)*15)

	sb.WriteString(`
        INSERT INTO invoice_items(
            invoice_id,
            description,
            qty,
            unit_price,
            discount_rate,
            discount_type,
            tax_rate,
            tax_type,
            discount_amount_per_unit,
            taxable_base_per_unit,
            tax_amount_per_unit,
            line_gross,
            line_discount,
            line_net,
            line_tax,
            line_total
        ) VALUES `,
	)

	for idx := range items {
		if idx > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString(`
            (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `)

		item := items[idx]

		args = append(
			args,
			i.ID().String(),
			item.Description(),
			item.Qty().String(),
			item.UnitPrice().String(),
			item.DiscountRate().String(),
			item.DiscountType().String(),
			item.TaxRate().String(),
			item.TaxType().String(),
			item.DiscountAmountPerUnit().String(),
			item.TaxableBasePerUnit().String(),
			item.TaxAmountPerUnit().String(),
			item.LineGross().String(),
			item.LineDiscount().String(),
			item.LineNet().String(),
			item.LineTax().String(),
			item.LineTotal().String(),
		)
	}

	_, err = ex.ExecContext(
		ctx,
		sb.String(),
		args...,
	)
	if err != nil {
		return fmt.Errorf(
			"billing.BillingRepository.Add: %w",
			err,
		)
	}

	return nil
}

func (r *BillingRepository) Get(ctx context.Context, id billing.InvoiceID) (*billing.Invoice, error) {

	ex := r.executor(ctx)

	var (
		invoiceID            string
		projectID            string
		isInvoice            bool
		status               string
		dueAt                *time.Time
		currencyCode         string
		currencyName         string
		currencySymbol       string
		currencyDecimalPlace int
		note                 *string
		totalTaxStr          string
		totalDiscountStr     string
		subTotalStr          string
		version              int
		createdAt            time.Time
		updatedAt            time.Time
	)

	err := ex.QueryRowContext(
		ctx,
		`
        SELECT
            i.id,
            i.project_id,
            i.is_invoice,
			i.status,
            i.due_at,
            i.currency_code,
            c.name,
            c.symbol,
            c.decimal_places,
            i.note,
            i.total_tax,
            i.total_discount,
            i.sub_total,
            i.version,
            i.created_at,
            i.updated_at
        FROM invoices i
        INNER JOIN currencies c
            ON c.code = i.currency_code
        WHERE i.id = ?
        `,
		id.String(),
	).Scan(
		&invoiceID,
		&projectID,
		&isInvoice,
		&status,
		&dueAt,
		&currencyCode,
		&currencyName,
		&currencySymbol,
		&currencyDecimalPlace,
		&note,
		&totalTaxStr,
		&totalDiscountStr,
		&subTotalStr,
		&version,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf(
				"%w: %w", billing.ErrBillingInvoiceNotFound,
				err,
			)
		}

		return nil, fmt.Errorf(
			"billing.BillingRepository.Get: %w",
			err,
		)
	}

	rows, err := ex.QueryContext(
		ctx,
		`
        SELECT
            description,
            qty,
            unit_price,
            discount_rate,
            discount_type,
            tax_rate,
            tax_type,
            discount_amount_per_unit,
            taxable_base_per_unit,
            tax_amount_per_unit,
            line_discount,
            line_tax,
            line_gross,
            line_net,
            line_total
        FROM invoice_items
        WHERE invoice_id = ?
        `,
		id.String(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.Get: %w",
			err,
		)
	}
	defer rows.Close()

	items := make([]billing.Item, 0)

	for rows.Next() {
		var (
			description              string
			qtyStr                   string
			unitPriceStr             string
			discountRateStr          string
			discountTypeStr          string
			taxRateStr               string
			taxTypeStr               string
			discountAmountPerUnitStr string
			taxableBasePerUnitStr    string
			taxAmountPerUnitStr      string
			lineDiscountStr          string
			lineTaxStr               string
			lineGrossStr             string
			lineNetStr               string
			lineTotalStr             string
		)

		err := rows.Scan(
			&description,
			&qtyStr,
			&unitPriceStr,
			&discountRateStr,
			&discountTypeStr,
			&taxRateStr,
			&taxTypeStr,
			&discountAmountPerUnitStr,
			&taxableBasePerUnitStr,
			&taxAmountPerUnitStr,
			&lineDiscountStr,
			&lineTaxStr,
			&lineGrossStr,
			&lineNetStr,
			&lineTotalStr,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: %w",
				err,
			)
		}

		qty, err := billing.NewQty(qtyStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid qty: %w",
				err,
			)
		}

		unitPrice, err := billing.NewDecimal(unitPriceStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid unit price: %w",
				err,
			)
		}

		discountRate, err := billing.NewDecimal(discountRateStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid discount rate: %w",
				err,
			)
		}

		taxRate, err := billing.NewDecimal(taxRateStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid tax rate: %w",
				err,
			)
		}

		discountAmountPerUnit, err := billing.NewDecimal(discountAmountPerUnitStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid discount amount per unit: %w",
				err,
			)
		}

		taxableBasePerUnit, err := billing.NewDecimal(taxableBasePerUnitStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid taxable base per unit: %w",
				err,
			)
		}

		taxAmountPerUnit, err := billing.NewDecimal(taxAmountPerUnitStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid tax amount per unit: %w",
				err,
			)
		}

		lineDiscount, err := billing.NewDecimal(lineDiscountStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid line discount: %w",
				err,
			)
		}

		lineTax, err := billing.NewDecimal(lineTaxStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid line tax: %w",
				err,
			)
		}

		lineGross, err := billing.NewDecimal(lineGrossStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid line gross: %w",
				err,
			)
		}

		lineNet, err := billing.NewDecimal(lineNetStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid line net: %w",
				err,
			)
		}

		lineTotal, err := billing.NewDecimal(lineTotalStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid line total: %w",
				err,
			)
		}

		items = append(items, billing.RestoreItem(
			description,
			qty,
			unitPrice,
			discountRate,
			billing.DiscountType(discountTypeStr),
			taxRate,
			billing.TaxType(taxTypeStr),
			discountAmountPerUnit,
			taxableBasePerUnit,
			taxAmountPerUnit,
			lineDiscount,
			lineTax,
			lineGross,
			lineNet,
			lineTotal,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.Get: %w",
			err,
		)
	}

	cCode, err := billing.NewCurrencyCode(currencyCode)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.Get: invalid currency code: %w",
			err,
		)
	}

	cName, err := billing.NewCurrencyName(currencyName)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.Get: invalid currency name: %w",
			err,
		)
	}

	cSymbol, err := billing.NewCurrencySymbol(currencySymbol)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.Get: invalid currency symbol: %w",
			err,
		)
	}

	cDecimals, err := billing.NewCurrencyDecimals(currencyDecimalPlace)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.Get: invalid currency decimal places: %w",
			err,
		)
	}

	currency := billing.NewCurrency(
		cName,
		cSymbol,
		cCode,
		cDecimals,
	)

	totalTax, err := billing.NewDecimal(totalTaxStr)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.Get: invalid total tax: %w",
			err,
		)
	}

	totalDiscount, err := billing.NewDecimal(totalDiscountStr)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.Get: invalid total discount: %w",
			err,
		)
	}

	subTotal, err := billing.NewDecimal(subTotalStr)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.Get: invalid subtotal: %w",
			err,
		)
	}

	projectIDValue, err := project.NewProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.Get: invalid project id: %w",
			err,
		)
	}

	statusValue, err := billing.NewStatus(status)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.Get: invalid status: %w",
			err,
		)
	}

	invoice := billing.RestoreInvoice(
		id,
		projectIDValue,
		isInvoice,
		statusValue,
		dueAt,
		currency,
		note,
		items,
		totalTax,
		totalDiscount,
		subTotal,
		version,
		createdAt,
		updatedAt,
	)

	return &invoice, nil
}

func (r *BillingRepository) GetCurrencyByCode(ctx context.Context, c billing.CurrencyCode) (billing.Currency, error) {
	ex := r.executor(ctx)

	var (
		code          string
		name          string
		symbol        string
		decimalPlaces int
	)

	err := ex.QueryRowContext(
		ctx,
		`
        SELECT
            code,
			name,
			symbol,
            decimal_places
        FROM currencies
        WHERE code = ?
        `,
		c.String(),
	).Scan(
		&code,
		&name,
		&symbol,
		&decimalPlaces,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return billing.Currency{}, fmt.Errorf(
				"%w: %w", billing.ErrBillingCurrencyNotFound, err,
			)
		}

		return billing.Currency{}, fmt.Errorf(
			"billing.BillingRepository.GetCurrencyByCode: %w",
			err,
		)
	}

	cCode, err := billing.NewCurrencyCode(code)
	if err != nil {
		return billing.Currency{}, fmt.Errorf(
			"billing.BillingRepository.GetCurrencyByCode: invalid currency code: %w",
			err,
		)
	}

	cName, err := billing.NewCurrencyName(name)
	if err != nil {
		return billing.Currency{}, fmt.Errorf(
			"billing.BillingRepository.Get: invalid currency name: %w",
			err,
		)
	}

	cSymbol, err := billing.NewCurrencySymbol(symbol)
	if err != nil {
		return billing.Currency{}, fmt.Errorf(
			"billing.BillingRepository.Get: invalid currency symbol: %w",
			err,
		)
	}

	cDecimals, err := billing.NewCurrencyDecimals(decimalPlaces)
	if err != nil {
		return billing.Currency{}, fmt.Errorf(
			"billing.BillingRepository.Get: invalid currency decimal places: %w",
			err,
		)
	}

	currency := billing.NewCurrency(
		cName,
		cSymbol,
		cCode,
		cDecimals,
	)

	return currency, nil
}

func (r *BillingRepository) GetBillingOverviewByMember(ctx context.Context, mID iam.UserID) (billing.BillingOverview, error) {
	ex := r.executor(ctx)

	rows, err := ex.QueryContext(
		ctx,
		`
        SELECT
            i.is_invoice,
            i.status,
            i.currency_code,
            COUNT(*) as total_count,
            COALESCE(SUM(i.sub_total), 0) as total_amount
        FROM invoices i
        INNER JOIN project_members pm
            ON pm.project_id = i.project_id
        WHERE pm.user_id = ?
        GROUP BY
            i.is_invoice,
            i.status,
            i.currency_code
        `,
		mID.String(),
	)
	if err != nil {
		return billing.BillingOverview{}, fmt.Errorf(
			"billing.BillingRepository.GetBillingOverviewByMember: %w",
			err,
		)
	}
	defer rows.Close()

	var (
		invoicesPaid      []billing.BillingOverviewMetric
		invoicesPending   []billing.BillingOverviewMetric
		invoicesAccepted  []billing.BillingOverviewMetric
		invoicesRejected  []billing.BillingOverviewMetric
		invoicesCancelled []billing.BillingOverviewMetric
		quotesPending     []billing.BillingOverviewMetric
		quotesRejected    []billing.BillingOverviewMetric
	)

	for rows.Next() {
		var (
			isInvoice   bool
			status      string
			currencyStr string
			count       int64
			amountStr   string
		)

		err := rows.Scan(
			&isInvoice,
			&status,
			&currencyStr,
			&count,
			&amountStr,
		)
		if err != nil {
			return billing.BillingOverview{}, fmt.Errorf(
				"billing.BillingRepository.GetBillingOverviewByMember: %w",
				err,
			)
		}

		currencyCode, err := billing.NewCurrencyCode(currencyStr)
		if err != nil {
			return billing.BillingOverview{}, fmt.Errorf(
				"billing.BillingRepository.GetBillingOverviewByMember: invalid currency code: %w",
				err,
			)
		}

		amount, err := billing.NewDecimal(amountStr)
		if err != nil {
			return billing.BillingOverview{}, fmt.Errorf(
				"billing.BillingRepository.GetBillingOverviewByMember: invalid amount: %w",
				err,
			)
		}

		metric := billing.NewBillingOverviewMetric(
			currencyCode,
			amount,
			count,
		)

		switch {
		case isInvoice && status == billing.StatusPaid.String():
			invoicesPaid = append(invoicesPaid, metric)

		case isInvoice && status == billing.StatusPending.String():
			invoicesPending = append(invoicesPending, metric)

		case isInvoice && status == billing.StatusAccepted.String():
			invoicesAccepted = append(invoicesAccepted, metric)

		case isInvoice && status == billing.StatusRejected.String():
			invoicesRejected = append(invoicesRejected, metric)

		case isInvoice && status == billing.StatusCancelled.String():
			invoicesCancelled = append(invoicesCancelled, metric)

		case !isInvoice && status == billing.StatusPending.String():
			quotesPending = append(quotesPending, metric)

		case !isInvoice && status == billing.StatusRejected.String():
			quotesRejected = append(quotesRejected, metric)
		}
	}

	if err := rows.Err(); err != nil {
		return billing.BillingOverview{}, fmt.Errorf(
			"billing.BillingRepository.GetBillingOverviewByMember: %w",
			err,
		)
	}

	return billing.NewBillingOverview(
		invoicesPaid,
		invoicesPending,
		invoicesAccepted,
		invoicesRejected,
		invoicesCancelled,
		quotesPending,
		quotesRejected,
	), nil
}

func (r *BillingRepository) GetStatsByProject(ctx context.Context, pID project.ProjectID) (billing.Stats, error) {
	ex := r.executor(ctx)

	var (
		invoiceCount     int
		invoicePaidCount int
		quoteCount       int
	)

	err := ex.QueryRowContext(
		ctx,
		`
		SELECT
			COUNT(*) FILTER (
				WHERE i.is_invoice = 1
			) AS invoice_count,

			COUNT(*) FILTER (
				WHERE i.is_invoice = 1
				AND i.status = ?
			) AS invoice_paid_count,

			COUNT(*) FILTER (
				WHERE i.is_invoice = 0
			) AS quote_count
		FROM invoices i
		WHERE i.project_id = ?
		`,
		billing.StatusPaid.String(),
		pID.String(),
	).Scan(
		&invoiceCount,
		&invoicePaidCount,
		&quoteCount,
	)
	if err != nil {
		return billing.Stats{}, fmt.Errorf(
			"billing.BillingRepository.GetStatsByProject: %w",
			err,
		)
	}

	stats := billing.NewStats(invoiceCount, invoicePaidCount, quoteCount)

	return stats, nil
}

func (r *BillingRepository) ListByMember(ctx context.Context, f billing.FilterByMember, p common.Page) ([]*billing.Invoice, error) {

	ex := r.executor(ctx)

	type invoiceRow struct {
		invoiceID            string
		projectID            string
		isInvoice            bool
		status               string
		dueAt                *time.Time
		currencyCode         string
		currencyName         string
		currencySymbol       string
		currencyDecimalPlace int
		note                 *string
		totalTaxStr          string
		totalDiscountStr     string
		subTotalStr          string
		version              int
		createdAt            time.Time
		updatedAt            time.Time
	}

	invoiceRows := make([]invoiceRow, 0)
	invoiceIDs := make([]string, 0)

	queryArgs := make([]any, 0)

	var invQuery strings.Builder

	invQuery.WriteString(`
    SELECT
        i.id,
        i.project_id,
        i.is_invoice,
        i.status,
        i.due_at,
        i.currency_code,
        c.name,
        c.symbol,
        c.decimal_places,
        i.note,
        i.total_tax,
        i.total_discount,
        i.sub_total,
        i.version,
        i.created_at,
        i.updated_at
    FROM invoices i
    INNER JOIN currencies c
        ON c.code = i.currency_code
    INNER JOIN project_members pm
        ON pm.project_id = i.project_id
    WHERE pm.user_id = ?
	`)

	queryArgs = append(queryArgs, f.MemberID.String())

	if f.Keyword != nil && strings.TrimSpace(*f.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(*f.Keyword) + "%"

		invQuery.WriteString(`
        AND (
            i.id LIKE ?
            OR i.note LIKE ?
        )
    `)

		queryArgs = append(
			queryArgs,
			keyword,
			keyword,
		)
	}

	if f.IsInvoice != nil {
		invQuery.WriteString(`
        AND i.is_invoice = ?
    `)

		queryArgs = append(
			queryArgs,
			*f.IsInvoice,
		)
	}

	if f.Status != nil {
		invQuery.WriteString(`
        AND i.status = ?
    `)

		queryArgs = append(
			queryArgs,
			string(*f.Status),
		)
	}

	invQuery.WriteString(`
    ORDER BY i.created_at DESC
    LIMIT ? OFFSET ?
	`)

	queryArgs = append(queryArgs, p.Limit(), p.Offset())

	rows, err := ex.QueryContext(ctx, invQuery.String(), queryArgs...)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListByMember: %w",
			err,
		)
	}
	defer rows.Close()

	for rows.Next() {
		var row invoiceRow

		err := rows.Scan(
			&row.invoiceID,
			&row.projectID,
			&row.isInvoice,
			&row.status,
			&row.dueAt,
			&row.currencyCode,
			&row.currencyName,
			&row.currencySymbol,
			&row.currencyDecimalPlace,
			&row.note,
			&row.totalTaxStr,
			&row.totalDiscountStr,
			&row.subTotalStr,
			&row.version,
			&row.createdAt,
			&row.updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: %w",
				err,
			)
		}

		invoiceRows = append(invoiceRows, row)
		invoiceIDs = append(invoiceIDs, row.invoiceID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListByMember: %w",
			err,
		)
	}

	if len(invoiceIDs) == 0 {
		return []*billing.Invoice{}, nil
	}

	args := make([]any, len(invoiceIDs))

	var b strings.Builder

	b.WriteString(`
        SELECT
            invoice_id,
            description,
            qty,
            unit_price,
            discount_rate,
            discount_type,
            tax_rate,
            tax_type,
            discount_amount_per_unit,
            taxable_base_per_unit,
            tax_amount_per_unit,
            line_discount,
            line_tax,
            line_gross,
            line_net,
            line_total
        FROM invoice_items
        WHERE invoice_id IN (
    `)

	for i, invoiceID := range invoiceIDs {
		if i > 0 {
			b.WriteString(",")
		}

		b.WriteString("?")
		args[i] = invoiceID
	}

	b.WriteString(")")

	itemRows, err := ex.QueryContext(
		ctx,
		b.String(),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListByMember: %w",
			err,
		)
	}
	defer itemRows.Close()

	itemsByInvoiceID := make(map[string][]billing.Item)

	for itemRows.Next() {
		var (
			invoiceID                string
			description              string
			qtyStr                   string
			unitPriceStr             string
			discountRateStr          string
			discountTypeStr          string
			taxRateStr               string
			taxTypeStr               string
			discountAmountPerUnitStr string
			taxableBasePerUnitStr    string
			taxAmountPerUnitStr      string
			lineDiscountStr          string
			lineTaxStr               string
			lineGrossStr             string
			lineNetStr               string
			lineTotalStr             string
		)

		err := itemRows.Scan(
			&invoiceID,
			&description,
			&qtyStr,
			&unitPriceStr,
			&discountRateStr,
			&discountTypeStr,
			&taxRateStr,
			&taxTypeStr,
			&discountAmountPerUnitStr,
			&taxableBasePerUnitStr,
			&taxAmountPerUnitStr,
			&lineDiscountStr,
			&lineTaxStr,
			&lineGrossStr,
			&lineNetStr,
			&lineTotalStr,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: %w",
				err,
			)
		}

		qty, err := billing.NewQty(qtyStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid qty: %w",
				err,
			)
		}

		unitPrice, err := billing.NewDecimal(unitPriceStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid unit price: %w",
				err,
			)
		}

		discountRate, err := billing.NewDecimal(discountRateStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid discount rate: %w",
				err,
			)
		}

		taxRate, err := billing.NewDecimal(taxRateStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid tax rate: %w",
				err,
			)
		}

		discountAmountPerUnit, err := billing.NewDecimal(discountAmountPerUnitStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid discount amount per unit: %w",
				err,
			)
		}

		taxableBasePerUnit, err := billing.NewDecimal(taxableBasePerUnitStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid taxable base per unit: %w",
				err,
			)
		}

		taxAmountPerUnit, err := billing.NewDecimal(taxAmountPerUnitStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid tax amount per unit: %w",
				err,
			)
		}

		lineDiscount, err := billing.NewDecimal(lineDiscountStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid line discount: %w",
				err,
			)
		}

		lineTax, err := billing.NewDecimal(lineTaxStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid line tax: %w",
				err,
			)
		}

		lineGross, err := billing.NewDecimal(lineGrossStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid line gross: %w",
				err,
			)
		}

		lineNet, err := billing.NewDecimal(lineNetStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid line net: %w",
				err,
			)
		}

		lineTotal, err := billing.NewDecimal(lineTotalStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid line total: %w",
				err,
			)
		}

		item := billing.RestoreItem(
			description,
			qty,
			unitPrice,
			discountRate,
			billing.DiscountType(discountTypeStr),
			taxRate,
			billing.TaxType(taxTypeStr),
			discountAmountPerUnit,
			taxableBasePerUnit,
			taxAmountPerUnit,
			lineDiscount,
			lineTax,
			lineGross,
			lineNet,
			lineTotal,
		)

		itemsByInvoiceID[invoiceID] = append(
			itemsByInvoiceID[invoiceID],
			item,
		)
	}

	if err := itemRows.Err(); err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListByMember: %w",
			err,
		)
	}

	invoices := make([]*billing.Invoice, 0, len(invoiceRows))

	for _, row := range invoiceRows {
		cCode, err := billing.NewCurrencyCode(row.currencyCode)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid currency code: %w",
				err,
			)
		}

		cName, err := billing.NewCurrencyName(row.currencyName)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid currency name: %w",
				err,
			)
		}

		cSymbol, err := billing.NewCurrencySymbol(row.currencySymbol)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid currency symbol: %w",
				err,
			)
		}

		cDecimals, err := billing.NewCurrencyDecimals(row.currencyDecimalPlace)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid currency decimal places: %w",
				err,
			)
		}

		currency := billing.NewCurrency(
			cName,
			cSymbol,
			cCode,
			cDecimals,
		)

		totalTax, err := billing.NewDecimal(row.totalTaxStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid total tax: %w",
				err,
			)
		}

		totalDiscount, err := billing.NewDecimal(row.totalDiscountStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid total discount: %w",
				err,
			)
		}

		subTotal, err := billing.NewDecimal(row.subTotalStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid subtotal: %w",
				err,
			)
		}

		projectIDValue, err := project.NewProjectID(row.projectID)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid project id: %w",
				err,
			)
		}

		status, err := billing.NewStatus(row.status)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid status: %w",
				err,
			)
		}

		invID, err := billing.NewInvoiceID(row.invoiceID)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByMember: invalid invoice id: %w",
				err,
			)
		}

		invoice := billing.RestoreInvoice(
			invID,
			projectIDValue,
			row.isInvoice,
			status,
			row.dueAt,
			currency,
			row.note,
			itemsByInvoiceID[row.invoiceID],
			totalTax,
			totalDiscount,
			subTotal,
			row.version,
			row.createdAt,
			row.updatedAt,
		)

		invoices = append(invoices, &invoice)
	}

	return invoices, nil
}

func (r *BillingRepository) ListByProject(ctx context.Context, f billing.FilterByProject, p common.Page) ([]*billing.Invoice, error) {
	ex := r.executor(ctx)

	type invoiceRow struct {
		invoiceID            string
		projectID            string
		isInvoice            bool
		status               string
		dueAt                *time.Time
		currencyCode         string
		currencyName         string
		currencySymbol       string
		currencyDecimalPlace int
		note                 *string
		totalTaxStr          string
		totalDiscountStr     string
		subTotalStr          string
		version              int
		createdAt            time.Time
		updatedAt            time.Time
	}

	invoiceRows := make([]invoiceRow, 0)
	invoiceIDs := make([]string, 0)

	queryArgs := make([]any, 0)

	var invQuery strings.Builder

	invQuery.WriteString(`
    SELECT
        i.id,
        i.project_id,
        i.is_invoice,
        i.status,
        i.due_at,
        i.currency_code,
        c.name,
        c.symbol,
        c.decimal_places,
        i.note,
        i.total_tax,
        i.total_discount,
        i.sub_total,
        i.version,
        i.created_at,
        i.updated_at
    FROM invoices i
    INNER JOIN currencies c
        ON c.code = i.currency_code
    WHERE i.project_id = ?
	`)

	queryArgs = append(queryArgs, f.ProjectID.String())

	if f.Keyword != nil && strings.TrimSpace(*f.Keyword) != "" {
		invQuery.WriteString(`
        AND (
            i.note LIKE ?
            OR i.id LIKE ?
        )
    `)

		keyword := "%" + strings.TrimSpace(*f.Keyword) + "%"
		queryArgs = append(queryArgs, keyword, keyword)
	}

	if f.IsInvoice != nil {
		invQuery.WriteString(`
        AND i.is_invoice = ?
    `)

		queryArgs = append(queryArgs, *f.IsInvoice)
	}

	if f.Status != nil {
		invQuery.WriteString(`
        AND i.status = ?
    `)

		queryArgs = append(queryArgs, string(*f.Status))
	}

	invQuery.WriteString(`
    ORDER BY i.created_at DESC
    LIMIT ? OFFSET ?
	`)

	queryArgs = append(queryArgs, p.Limit(), p.Offset())
	rows, err := ex.QueryContext(ctx, invQuery.String(), queryArgs...)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListByProject: %w",
			err,
		)
	}
	defer rows.Close()

	for rows.Next() {
		var row invoiceRow

		err := rows.Scan(
			&row.invoiceID,
			&row.projectID,
			&row.isInvoice,
			&row.status,
			&row.dueAt,
			&row.currencyCode,
			&row.currencyName,
			&row.currencySymbol,
			&row.currencyDecimalPlace,
			&row.note,
			&row.totalTaxStr,
			&row.totalDiscountStr,
			&row.subTotalStr,
			&row.version,
			&row.createdAt,
			&row.updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: %w",
				err,
			)
		}

		invoiceRows = append(invoiceRows, row)
		invoiceIDs = append(invoiceIDs, row.invoiceID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListByProject: %w",
			err,
		)
	}

	if len(invoiceIDs) == 0 {
		return []*billing.Invoice{}, nil
	}

	args := make([]any, len(invoiceIDs))

	var itemsQuery strings.Builder

	itemsQuery.WriteString(`
        SELECT
            invoice_id,
            description,
            qty,
            unit_price,
            discount_rate,
            discount_type,
            tax_rate,
            tax_type,
            discount_amount_per_unit,
            taxable_base_per_unit,
            tax_amount_per_unit,
            line_discount,
            line_tax,
            line_gross,
            line_net,
            line_total
        FROM invoice_items
        WHERE invoice_id IN (
    `)

	for i, invoiceID := range invoiceIDs {
		if i > 0 {
			itemsQuery.WriteString(",")
		}

		itemsQuery.WriteString("?")
		args[i] = invoiceID
	}

	itemsQuery.WriteString(")")

	itemRows, err := ex.QueryContext(
		ctx,
		itemsQuery.String(),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListByProject: %w",
			err,
		)
	}
	defer itemRows.Close()

	itemsByInvoiceID := make(map[string][]billing.Item)

	for itemRows.Next() {
		var (
			invoiceID                string
			description              string
			qtyStr                   string
			unitPriceStr             string
			discountRateStr          string
			discountTypeStr          string
			taxRateStr               string
			taxTypeStr               string
			discountAmountPerUnitStr string
			taxableBasePerUnitStr    string
			taxAmountPerUnitStr      string
			lineDiscountStr          string
			lineTaxStr               string
			lineGrossStr             string
			lineNetStr               string
			lineTotalStr             string
		)

		err := itemRows.Scan(
			&invoiceID,
			&description,
			&qtyStr,
			&unitPriceStr,
			&discountRateStr,
			&discountTypeStr,
			&taxRateStr,
			&taxTypeStr,
			&discountAmountPerUnitStr,
			&taxableBasePerUnitStr,
			&taxAmountPerUnitStr,
			&lineDiscountStr,
			&lineTaxStr,
			&lineGrossStr,
			&lineNetStr,
			&lineTotalStr,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: %w",
				err,
			)
		}

		qty, err := billing.NewQty(qtyStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid qty: %w",
				err,
			)
		}

		unitPrice, err := billing.NewDecimal(unitPriceStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid unit price: %w",
				err,
			)
		}

		discountRate, err := billing.NewDecimal(discountRateStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid discount rate: %w",
				err,
			)
		}

		taxRate, err := billing.NewDecimal(taxRateStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid tax rate: %w",
				err,
			)
		}

		discountAmountPerUnit, err := billing.NewDecimal(discountAmountPerUnitStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid discount amount per unit: %w",
				err,
			)
		}

		taxableBasePerUnit, err := billing.NewDecimal(taxableBasePerUnitStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid taxable base per unit: %w",
				err,
			)
		}

		taxAmountPerUnit, err := billing.NewDecimal(taxAmountPerUnitStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid tax amount per unit: %w",
				err,
			)
		}

		lineDiscount, err := billing.NewDecimal(lineDiscountStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid line discount: %w",
				err,
			)
		}

		lineTax, err := billing.NewDecimal(lineTaxStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid line tax: %w",
				err,
			)
		}

		lineGross, err := billing.NewDecimal(lineGrossStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid line gross: %w",
				err,
			)
		}

		lineNet, err := billing.NewDecimal(lineNetStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid line net: %w",
				err,
			)
		}

		lineTotal, err := billing.NewDecimal(lineTotalStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid line total: %w",
				err,
			)
		}

		item := billing.RestoreItem(
			description,
			qty,
			unitPrice,
			discountRate,
			billing.DiscountType(discountTypeStr),
			taxRate,
			billing.TaxType(taxTypeStr),
			discountAmountPerUnit,
			taxableBasePerUnit,
			taxAmountPerUnit,
			lineDiscount,
			lineTax,
			lineGross,
			lineNet,
			lineTotal,
		)

		itemsByInvoiceID[invoiceID] = append(
			itemsByInvoiceID[invoiceID],
			item,
		)
	}

	if err := itemRows.Err(); err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListByProject: %w",
			err,
		)
	}

	invoices := make([]*billing.Invoice, 0, len(invoiceRows))

	for _, row := range invoiceRows {
		cCode, err := billing.NewCurrencyCode(row.currencyCode)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid currency code: %w",
				err,
			)
		}

		cName, err := billing.NewCurrencyName(row.currencyName)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid currency name: %w",
				err,
			)
		}

		cSymbol, err := billing.NewCurrencySymbol(row.currencySymbol)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid currency symbol: %w",
				err,
			)
		}

		cDecimals, err := billing.NewCurrencyDecimals(row.currencyDecimalPlace)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.Get: invalid currency decimal places: %w",
				err,
			)
		}

		currency := billing.NewCurrency(
			cName,
			cSymbol,
			cCode,
			cDecimals,
		)

		totalTax, err := billing.NewDecimal(row.totalTaxStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid total tax: %w",
				err,
			)
		}

		totalDiscount, err := billing.NewDecimal(row.totalDiscountStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid total discount: %w",
				err,
			)
		}

		subTotal, err := billing.NewDecimal(row.subTotalStr)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid subtotal: %w",
				err,
			)
		}

		projectIDValue, err := project.NewProjectID(row.projectID)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid project id: %w",
				err,
			)
		}

		status, err := billing.NewStatus(row.status)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid status: %w",
				err,
			)
		}

		invID, err := billing.NewInvoiceID(row.invoiceID)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListByProject: invalid invoice id: %w",
				err,
			)
		}

		invoice := billing.RestoreInvoice(
			invID,
			projectIDValue,
			row.isInvoice,
			status,
			row.dueAt,
			currency,
			row.note,
			itemsByInvoiceID[row.invoiceID],
			totalTax,
			totalDiscount,
			subTotal,
			row.version,
			row.createdAt,
			row.updatedAt,
		)

		invoices = append(invoices, &invoice)
	}

	return invoices, nil
}

func (r *BillingRepository) ListMonthlyPaidCountByProject(ctx context.Context, pID project.ProjectID) ([]billing.Metric, error) {
	ex := r.executor(ctx)

	query := `
		WITH RECURSIVE months(year_month, start_date) AS (
			SELECT
				strftime('%Y-%m', 'now'),
				date(strftime('%Y-%m', 'now') || '-01')

			UNION ALL

			SELECT
				strftime('%Y-%m', date(start_date, '-1 month')),
				date(start_date, '-1 month')
			FROM months
			WHERE start_date > date('now', '-11 months')
		)
		SELECT
			m.year_month,
			COUNT(i.id) AS paid_count
		FROM months m
		LEFT JOIN invoices i
			ON strftime('%Y-%m', i.updated_at) = m.year_month
			AND i.project_id = ?
			AND i.status = ?
		GROUP BY m.year_month
		ORDER BY m.year_month ASC
	`

	rows, err := ex.QueryContext(
		ctx,
		query,
		pID.String(),
		billing.StatusPaid.String(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListPaidCountByProject: %w",
			err,
		)
	}
	defer rows.Close()

	metrics := make([]billing.Metric, 0, 12)

	for rows.Next() {
		var (
			key   string
			value int64
		)

		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListPaidCountByProject: %w",
				err,
			)
		}

		metrics = append(
			metrics,
			billing.NewMetric(key, value),
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListPaidCountByProject: %w",
			err,
		)
	}

	return metrics, nil
}

func (r *BillingRepository) ListMonthlyPaidCountByMember(ctx context.Context, mID iam.UserID) ([]billing.Metric, error) {
	ex := r.executor(ctx)

	query := `
		WITH RECURSIVE months(year_month, start_date) AS (
			SELECT
				strftime('%Y-%m', 'now'),
				date(strftime('%Y-%m', 'now') || '-01')

			UNION ALL

			SELECT
				strftime('%Y-%m', date(start_date, '-1 month')),
				date(start_date, '-1 month')
			FROM months
			WHERE start_date > date('now', '-11 months')
		)
		SELECT
			m.year_month,
			COUNT(i.id) AS paid_count
		FROM months m
		LEFT JOIN invoices i
			ON strftime('%Y-%m', i.updated_at) = m.year_month
			AND i.status = ?
			AND i.project_id IN (
				SELECT project_id
				FROM project_members
				WHERE user_id = ?
			)
		GROUP BY m.year_month
		ORDER BY m.year_month ASC
	`

	rows, err := ex.QueryContext(
		ctx,
		query,
		billing.StatusPaid.String(),
		mID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListMonthlyPaidCountByMember: %w",
			err,
		)
	}
	defer rows.Close()

	metrics := make([]billing.Metric, 0, 12)

	for rows.Next() {
		var (
			key   string
			value int64
		)

		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListMonthlyPaidCountByMember: %w",
				err,
			)
		}

		metrics = append(metrics, billing.NewMetric(key, value))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListMonthlyPaidCountByMember: %w",
			err,
		)
	}

	return metrics, nil
}

func (r *BillingRepository) ListStatsByProjects(ctx context.Context, projectIDs []project.ProjectID) (map[project.ProjectID]billing.Stats, error) {
	ex := r.executor(ctx)

	statsByProjects := make(map[project.ProjectID]billing.Stats, len(projectIDs))

	if len(projectIDs) == 0 {
		return statsByProjects, nil
	}

	var query strings.Builder

	query.WriteString(`
		SELECT
			project_id,
			COUNT(*) AS invoice_count,
			COUNT(*) FILTER ( WHERE status = ? ) AS invoice_paid_count,
			COUNT(*) FILTER ( WHERE is_invoice = 0 ) AS quote_count
		FROM invoices
		WHERE project_id IN (
	`)

	args := make([]any, 0, len(projectIDs)+1)
	args = append(args, billing.StatusPaid.String())

	for i := range projectIDs {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString("?")
		args = append(args, projectIDs[i].String())
	}

	query.WriteString(`
		)
		GROUP BY project_id
	`)

	rows, err := ex.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListStatsByProjects: %w",
			err,
		)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			rawProjectID     string
			invoiceCount     int
			invoicePaidCount int
			quoteCount       int
		)

		if err := rows.Scan(
			&rawProjectID,
			&invoiceCount,
			&invoicePaidCount,
			&quoteCount,
		); err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListStatsByProjects: %w",
				err,
			)
		}

		pID, err := project.NewProjectID(rawProjectID)
		if err != nil {
			return nil, fmt.Errorf(
				"billing.BillingRepository.ListStatsByProjects: %w",
				err,
			)
		}

		statsByProjects[pID] = billing.NewStats(
			invoiceCount,
			invoicePaidCount,
			quoteCount,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"billing.BillingRepository.ListStatsByProjects: %w",
			err,
		)
	}

	return statsByProjects, nil
}

func (r *BillingRepository) CountByProject(ctx context.Context, f billing.FilterByProject) (int, error) {
	ex := r.executor(ctx)

	var query strings.Builder

	query.WriteString(`
		SELECT
			COUNT(*)
		FROM invoices i
		INNER JOIN projects p ON p.id = i.project_id
		WHERE i.project_id = ?
	`)

	args := make([]any, 0, 5)
	args = append(args, f.ProjectID.String())

	if f.Keyword != nil && *f.Keyword != "" {
		query.WriteString(" AND ( p.name LIKE ? OR i.note LIKE ? )")
		kw := "%" + *f.Keyword + "%"
		args = append(args, kw, kw)
	}

	if f.IsInvoice != nil {
		query.WriteString(" AND i.is_invoice = ?")
		args = append(args, *f.IsInvoice)
	}

	if f.Status != nil {
		query.WriteString(" AND i.status = ?")
		args = append(args, f.Status.String())
	}

	var count int
	err := ex.QueryRowContext(ctx, query.String(), args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("billing.BillingRepository.CountByProject: %w", err)
	}

	return count, nil
}

func (r *BillingRepository) CountByMember(ctx context.Context, f billing.FilterByMember) (int, error) {
	ex := r.executor(ctx)

	var query strings.Builder

	query.WriteString(`
		SELECT
			COUNT(*)
		FROM invoices i
		INNER JOIN projects p ON p.id = i.project_id
		INNER JOIN project_members pm ON pm.project_id = i.project_id
		WHERE pm.user_id = ?
	`)

	args := make([]any, 0, 5)
	args = append(args, f.MemberID.String())

	if f.Keyword != nil && *f.Keyword != "" {
		query.WriteString(" AND ( p.name LIKE ? OR i.note LIKE ? )")
		kw := "%" + *f.Keyword + "%"
		args = append(args, kw, kw)
	}

	if f.IsInvoice != nil {
		query.WriteString(" AND i.is_invoice = ?")
		args = append(args, *f.IsInvoice)
	}

	if f.Status != nil {
		query.WriteString(" AND i.status = ?")
		args = append(args, f.Status.String())
	}

	var count int
	err := ex.QueryRowContext(ctx, query.String(), args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("billing.BillingRepository.CountByMember: %w", err)
	}

	return count, nil
}

func (r *BillingRepository) Save(ctx context.Context, i *billing.Invoice) error {

	ex := r.executor(ctx)

	result, err := ex.ExecContext(
		ctx,
		`
		UPDATE invoices SET
			project_id = ?,
			is_invoice = ?,
			status = ?,
			due_at = ?,
			currency_code = ?,
			note = ?,
			total_tax = ?,
			total_discount = ?,
			sub_total = ?,
			version = version + 1,
			updated_at = ?
		WHERE id = ? AND version = ?
		`,
		i.ProjectID().String(),
		i.IsInvoice(),
		i.Status().String(),
		i.DueAt(),
		i.Currency().Code().String(),
		i.Note(),
		i.TotalTax().String(),
		i.TotalDiscount().String(),
		i.SubTotal().String(),
		i.UpdatedAt(),
		i.ID().String(),
		i.Version(),
	)
	if err != nil {
		return fmt.Errorf(
			"billing.BillingRepository.Save: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"billing.BillingRepository.Save: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(
			"%w: version mismatch",
			billing.ErrBillingConcurrentModification,
		)
	}

	_, err = ex.ExecContext(
		ctx,
		`DELETE FROM invoice_items WHERE invoice_id = ?`,
		i.ID().String(),
	)
	if err != nil {
		return fmt.Errorf("billing.BillingRepository.Save: %w", err)
	}

	items := i.Items()
	if len(items) == 0 {
		return nil
	}

	var sb strings.Builder
	args := make([]any, 0, len(items)*16)

	sb.WriteString(`
		INSERT INTO invoice_items(
			invoice_id,
			description,
			qty,
			unit_price,
			discount_rate,
			discount_type,
			tax_rate,
			tax_type,
			discount_amount_per_unit,
			taxable_base_per_unit,
			tax_amount_per_unit,
			line_gross,
			line_discount,
			line_net,
			line_tax,
			line_total
		) VALUES 
	`)

	for idx := range items {
		if idx > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString(`
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)

		it := items[idx]

		args = append(
			args,
			i.ID().String(),
			it.Description(),
			it.Qty().String(),
			it.UnitPrice().String(),
			it.DiscountRate().String(),
			it.DiscountType().String(),
			it.TaxRate().String(),
			it.TaxType().String(),
			it.DiscountAmountPerUnit().String(),
			it.TaxableBasePerUnit().String(),
			it.TaxAmountPerUnit().String(),
			it.LineGross().String(),
			it.LineDiscount().String(),
			it.LineNet().String(),
			it.LineTax().String(),
			it.LineTotal().String(),
		)
	}

	_, err = ex.ExecContext(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf("billing.BillingRepository.Save: %w", err)
	}

	return nil
}
