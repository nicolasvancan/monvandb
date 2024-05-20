package main

func main() {
	teste := map[string]string{
		"base":      "utils.Monvan_paths['base']",
		"databases": "utils.Monvan_paths[databases]",
		"users":     "utils.Monvan_paths[users]",
	}

	opa := []string{"base", "databases", "users"}

	for i, v := range opa {
		println(i, v, teste[v])
	}
}
