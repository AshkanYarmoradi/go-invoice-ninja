# Test Data

This directory contains sample JSON responses from the Invoice Ninja API.

## Structure

```
testdata/
├── fixtures/
│   ├── payment.json         # A single payment response
│   └── payments_list.json   # A list of payments with pagination metadata
└── README.md
```

## Usage

The test suite does not currently load these fixtures. Unit tests define their mock API responses inline in the test code.

To use a fixture in a new test, you could add a helper like the one below. It is only an example; the test suite does not provide it:

```go
func loadFixture(t *testing.T, name string) []byte {
    data, err := os.ReadFile(filepath.Join("testdata", "fixtures", name))
    if err != nil {
        t.Fatalf("Failed to load fixture %s: %v", name, err)
    }
    return data
}
```
