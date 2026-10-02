# Email Template Customization

Mizu includes built-in HTML email templates that are embedded into the application.

All email templates use Go's `html/template` syntax. Templates can contain static HTML as well as dynamic values provided by Mizu when the email is generated.

Custom templates can be provided using the `EMAIL_TEMPLATE_DIR` configuration option.

See the [Configuration Guide](configuration.md) for information about configuring `EMAIL_TEMPLATE_DIR`.

## Template Files

Mizu currently supports the following email templates:

| Template   | File            | Purpose                           |
| ---------- | --------------- | --------------------------------- |
| `verify`   | `verify.html`   | Account verification              |
| `verified` | `verified.html` | Account verification confirmation |
| `contract` | `contract.html` | Contract notification             |
| `invoice`  | `invoice.html`  | Invoice notification              |
| `recovery` | `recovery.html` | Account recovery                  |

Template names are derived from the filename without the `.html` extension.

For example:

```text
verify.html → verify
invoice.html → invoice
```

Custom templates must use the `.html` extension.

## Custom Template Directory

Set `EMAIL_TEMPLATE_DIR` to a directory containing custom templates:

```env
EMAIL_TEMPLATE_DIR=/path/to/templates
```

For example:

```text
templates/
├── verify.html
├── verified.html
├── contract.html
├── invoice.html
└── recovery.html
```

When Mizu starts, the built-in templates are loaded first.

If a custom template has the same name as a built-in template, the custom version replaces the built-in version.

For example:

```text
templates/
└── invoice.html
```

replaces the built-in `invoice` template while the other built-in templates remain unchanged.

Custom directories may also contain additional `.html` templates. These templates are loaded by Mizu, but they are only used if application code explicitly requests them by name.

Files that do not have the `.html` extension are ignored.

## Go Templates

Mizu uses Go's `html/template` package to parse and render email templates.

Dynamic values are inserted using Go template actions:

```html
<h1>Hello, {{.Name}}</h1>
```

A field is accessed from the data provided to the template using `.FieldName`.

For example:

```html
<a href="{{.BaseURL}}/verify?token={{.VerificationID}}">
  Verify your account
</a>
```

Template actions can also be used inside HTML attributes:

```html
<a href="{{.BaseURL}}/projects/{{.ProjectID}}"> Open Project </a>
```

Because Mizu uses `html/template`, dynamic values are HTML-escaped when rendered.

## Template Data

Each email type receives a specific set of data from the application.

### Account Verification Email

Template:

```text
verify.html
```

Available fields:

| Field             | Type     | Description                         |
| ----------------- | -------- | ----------------------------------- |
| `.BaseURL`        | `string` | Public URL of the Mizu installation |
| `.VerificationID` | `string` | Verification identifier             |

Example:

```html
<p>Verify your account by clicking the link below.</p>

<a href="{{.BaseURL}}/verify?token={{.VerificationID}}"> Verify Account </a>
```

### Account Verified Email

Template:

```text
verified.html
```

This template does not receive any dynamic data.

The template can therefore contain static HTML:

```html
<h1>Your account has been verified</h1>

<p>You can now sign in to Mizu.</p>
```

### Account Recovery Email

Template:

```text
recovery.html
```

Available fields:

| Field            | Type     | Description                         |
| ---------------- | -------- | ----------------------------------- |
| `.BaseURL`       | `string` | Public URL of the Mizu installation |
| `.RecoveryToken` | `string` | Account recovery token              |

Example:

```html
<p>A request was made to recover your account.</p>

<a href="{{.BaseURL}}/recovery?token={{.RecoveryToken}}"> Recover Account </a>
```

### Contract Email

Template:

```text
contract.html
```

Available fields:

| Field            | Type                         | Description                         |
| ---------------- | ---------------------------- | ----------------------------------- |
| `.BaseURL`       | `string`                     | Public URL of the Mizu installation |
| `.ContractID`    | `string`                     | Contract identifier                 |
| `.ContractName`  | `string`                     | Contract name                       |
| `.ContractTerms` | `string`                     | Contract terms                      |
| `.ProjectID`     | `string`                     | Project identifier                  |
| `.ProjectName`   | `string`                     | Project name                        |
| `.Signatories`   | `[]mailer.ContractSignatory` | Contract signatories                |

Example:

```html
<h1>{{.ContractName}}</h1>

<p>Project: {{.ProjectName}}</p>

<p>{{.ContractTerms}}</p>

<a href="{{.BaseURL}}/projects/{{.ProjectID}}"> Open Project </a>
```

#### Signatories

`.Signatories` is a collection of contract signatories.

It can be iterated over using the Go template `range` action:

```html
<h2>Signatories</h2>

<ul>
  {{range .Signatories}}
  <li>{{.FirstName}} {{.LastName}} — {{.Role}}</li>
  {{end}}
</ul>
```

Each signatory provides the following fields:

| Field        | Type        | Description              |
| ------------ | ----------- | ------------------------ |
| `.ID`        | `string`    | Signatory identifier     |
| `.FirstName` | `string`    | Signatory first name     |
| `.LastName`  | `string`    | Signatory last name      |
| `.Email`     | `string`    | Signatory email address  |
| `.Title`     | `*string`   | Optional signatory title |
| `.Role`      | `string`    | Signatory role           |
| `.Status`    | `string`    | Signatory status         |
| `.UpdatedAt` | `time.Time` | Last update timestamp    |

For example:

```html
{{range .Signatories}}
<div>
  <h3>{{.FirstName}} {{.LastName}}</h3>

  <p>Email: {{.Email}}</p>
  <p>Role: {{.Role}}</p>
  <p>Status: {{.Status}}</p>

  {{if .Title}}
  <p>Title: {{.Title}}</p>
  {{end}}
</div>
{{end}}
```

Because `.Title` is optional, it can be checked with `if` before rendering it.

`.UpdatedAt` is a `time.Time` value and can be accessed directly in a template:

```html
<p>Updated: {{.UpdatedAt}}</p>
```

### Invoice Email

Template:

```text
invoice.html
```

Available fields:

| Field                 | Type                   | Description                                        |
| --------------------- | ---------------------- | -------------------------------------------------- |
| `.BaseURL`            | `string`               | Public URL of the Mizu installation                |
| `.InvoiceID`          | `string`               | Full invoice identifier                            |
| `.FormattedInvoiceID` | `string`               | Short, formatted invoice identifier                |
| `.ProjectID`          | `string`               | Project identifier                                 |
| `.Status`             | `string`               | Invoice status                                     |
| `.CurrencyName`       | `string`               | Currency name                                      |
| `.CurrencyCode`       | `string`               | Currency code                                      |
| `.DueAt`              | `string`               | Formatted due date, or empty if no due date exists |
| `.Note`               | `*string`              | Optional invoice note                              |
| `.Items`              | `[]mailer.InvoiceItem` | Invoice line items                                 |
| `.SubTotal`           | `string`               | Invoice subtotal                                   |
| `.TotalTax`           | `string`               | Total tax                                          |
| `.TotalDiscount`      | `string`               | Total discount                                     |

Example:

```html
<h1>Invoice {{.FormattedInvoiceID}}</h1>

<p>Status: {{.Status}}</p>

<p>Currency: {{.CurrencyCode}}</p>

{{if .DueAt}}
<p>Due: {{.DueAt}}</p>
{{end}} {{if .Note}}
<p>{{.Note}}</p>
{{end}}

<h2>Items</h2>

<ul>
  {{range .Items}}
  <li>{{.Description}} — {{.LineTotal}}</li>
  {{end}}
</ul>

<p>Subtotal: {{.SubTotal}}</p>
<p>Tax: {{.TotalTax}}</p>
<p>Discount: {{.TotalDiscount}}</p>
```

#### Invoice Items

`.Items` is a collection of invoice line items.

Each item provides the following fields:

| Field           | Type     | Description          |
| --------------- | -------- | -------------------- |
| `.Description`  | `string` | Item description     |
| `.Qty`          | `string` | Quantity             |
| `.UnitPrice`    | `string` | Unit price           |
| `.DiscountRate` | `string` | Discount rate        |
| `.DiscountType` | `string` | Discount type        |
| `.TaxRate`      | `string` | Tax rate             |
| `.TaxType`      | `string` | Tax type             |
| `.LineGross`    | `string` | Gross line amount    |
| `.LineDiscount` | `string` | Line discount amount |
| `.LineNet`      | `string` | Net line amount      |
| `.LineTax`      | `string` | Line tax amount      |
| `.LineTotal`    | `string` | Final line total     |

For example:

```html
<table>
  <thead>
    <tr>
      <th>Description</th>
      <th>Quantity</th>
      <th>Unit Price</th>
      <th>Discount</th>
      <th>Tax</th>
      <th>Total</th>
    </tr>
  </thead>

  <tbody>
    {{range .Items}}
    <tr>
      <td>{{.Description}}</td>
      <td>{{.Qty}}</td>
      <td>{{.UnitPrice}}</td>
      <td>{{.LineDiscount}}</td>
      <td>{{.LineTax}}</td>
      <td>{{.LineTotal}}</td>
    </tr>
    {{end}}
  </tbody>
</table>
```

The individual pricing fields can also be displayed when required:

```html
{{range .Items}}
<div>
  <p>{{.Description}}</p>
  <p>Quantity: {{.Qty}}</p>
  <p>Unit price: {{.UnitPrice}}</p>

  <p>Discount: {{.DiscountRate}} {{.DiscountType}}</p>

  <p>Tax: {{.TaxRate}} {{.TaxType}}</p>

  <p>Gross: {{.LineGross}}</p>
  <p>Discount amount: {{.LineDiscount}}</p>
  <p>Net: {{.LineNet}}</p>
  <p>Tax amount: {{.LineTax}}</p>
  <p>Total: {{.LineTotal}}</p>
</div>
{{end}}
```

## Optional Values

Some template values may be optional.

For example, `.DueAt` is an empty string when an invoice does not have a due date:

```html
{{if .DueAt}}
<p>Due: {{.DueAt}}</p>
{{end}}
```

`.Note` is a pointer and can be checked using `if`:

```html
{{if .Note}}
<p>{{.Note}}</p>
{{end}}
```

`.Title` on a contract signatory is also optional:

```html
{{range .Signatories}} {{if .Title}}
<p>{{.Title}}</p>
{{end}} {{end}}
```

This allows templates to omit sections when optional data is unavailable.

## Go Template Control Structures

Go templates provide control structures that can be used to customize email content.

### Conditions

Use `if` to conditionally render content:

```html
{{if .Note}}
<p>{{.Note}}</p>
{{end}}
```

You can also provide an alternative using `else`:

```html
{{if .DueAt}}
<p>Due: {{.DueAt}}</p>
{{else}}
<p>No due date specified.</p>
{{end}}
```

### Iterating Over Collections

Use `range` to iterate over collections:

```html
<ul>
  {{range .Items}}
  <li>{{.Description}}</li>
  {{end}}
</ul>
```

The same pattern can be used with `.Signatories`:

```html
<ul>
  {{range .Signatories}}
  <li>{{.FirstName}} {{.LastName}}</li>
  {{end}}
</ul>
```

### Accessing Fields

Fields are accessed using a leading dot:

```html
{{.ProjectName}}
```

Nested fields can be accessed using additional dots when the underlying value provides them:

```html
{{.SomeObject.SomeField}}
```

Only fields and methods available on the value supplied by the application can be accessed.

## HTML Email

Templates are rendered as HTML emails with UTF-8 encoding:

```text
Content-Type: text/html; charset=UTF-8
```

Templates can therefore contain normal HTML markup:

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <title>Invoice</title>
  </head>

  <body>
    <h1>Invoice {{.FormattedInvoiceID}}</h1>
  </body>
</html>
```

For broad email-client compatibility, keep HTML and CSS relatively simple and test customized templates with the email clients you intend to support.

## Template Validation

Templates are parsed when Mizu starts.

If a custom template contains invalid Go template syntax, Mizu will fail to load the template and report the parsing error during startup.

For example, an invalid action such as:

```html
<h1>{{.InvoiceID</h1>
```

will cause template parsing to fail.

Always verify that Mizu starts successfully after adding or modifying custom templates.

## Security

Mizu uses Go's `html/template` package rather than `text/template`.

Dynamic values are therefore HTML-escaped when inserted into HTML templates.

Do not deliberately bypass this escaping when displaying user-controlled data.

Custom templates should also be treated as application code. Do not place passwords, API keys, database credentials, or other sensitive configuration values directly in template files.

## Customization Example

A minimal custom verification template might look like:

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <title>Verify Your Account</title>
  </head>

  <body>
    <h1>Welcome to Mizu</h1>

    <p>Thank you for creating an account.</p>

    <p>
      <a href="{{.BaseURL}}/verify?token={{.VerificationID}}">
        Verify your account
      </a>
    </p>

    <p>If you did not create this account, you can ignore this email.</p>
  </body>
</html>
```

Save it as:

```text
templates/verify.html
```

and configure:

```env
EMAIL_TEMPLATE_DIR=/path/to/templates
```

After restarting Mizu, the custom template will be used instead of the built-in verification template.

## Related Documentation

- [Configuration](configuration.md) — configuration options and environment variables
- [Development](development.md) — local development and testing
- [Deployment](deployment.md) — production deployment
