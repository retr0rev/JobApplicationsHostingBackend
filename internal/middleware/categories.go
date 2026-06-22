package middleware

// Categories is the canonical list of allowed job categories. New jobs
// must use one of these exactly (case-sensitive match). Clients do not
// create their own categories — see lib/categories.ts in each frontend.
var Categories = []string{
	"Engineering",
	"Design",
	"Product",
	"Marketing",
	"Sales",
	"Operations",
	"Customer Support",
	"Data",
	"Finance",
	"People",
	"Other",
}

// IsValidCategory reports whether cat is in the canonical list.
func IsValidCategory(cat string) bool {
	for _, c := range Categories {
		if c == cat {
			return true
		}
	}
	return false
}

// ValidateCategoryInList returns "" if valid, else an error message.
func ValidateCategoryInList(cat string) string {
	if cat == "" {
		return "category is required"
	}
	if !IsValidCategory(cat) {
		return "category must be one of the allowed values"
	}
	return ""
}
