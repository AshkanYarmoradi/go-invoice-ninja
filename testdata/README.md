# Test Data

This directory contains test fixtures and mock data for unit tests.

## Structure

```
testdata/
├── fixtures/           # JSON response fixtures
│   ├── payments/
│   ├── invoices/
│   └── clients/
└── README.md
```

## Usage

Test fixtures are used by the test suite to mock API responses:

```go
func loadFixture(t *testing.T, name string) []byte {
    data, err := os.ReadFile(filepath.Join("testdata", "fixtures", name))
    if err != nil {
        t.Fatalf("Failed to load fixture %s: %v", name, err)
    }
    return data
}
```
