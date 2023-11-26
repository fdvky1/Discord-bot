package core

import (
	supa "github.com/nedpals/supabase-go"
)

var Supabase *supa.Client

// func init() {
// 	Supabase = supa.CreateClient(
// 		os.Getenv("SUPABASE_URL"),
// 		os.Getenv("SUPABASE_KEY"),
// 	)
// }
