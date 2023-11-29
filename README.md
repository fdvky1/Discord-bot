
# Discord Bot

Backend for [Discord Bot frontend](https://github.com/fdvky1/discord-bot-fe)



## Setup

Clone this project\
Create project on [Supabase](https://supabase.com)\
Open SQL Editor on [Supabase](https://supabase.com) and [Execute this](https://gist.github.com/fdvky1/1bf95e80e2155c228e1ba050aa29ff35)

fill the .env

```bash
SUPABASE_URL="your project url"
SUPABASE_KEY="service role key(for bypass RLS)"
POSTGRESQL_URL="postgresql url(if you want to use postgre from supabase please enable the RLS after table created)"
```


    
## Run

```bash
  go mod tidy
  go build -o bot main.go
  ./bot
```