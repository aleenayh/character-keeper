var ctx = context.Background()

func main() {
  opt, _ := redis.ParseURL("rediss://default:********@one-pug-13888.upstash.io:6379")
  client := redis.NewClient(opt)

  client.Set(ctx, "foo", "bar", 0)
  val := client.Get(ctx, "foo").Val()
  print(val)
}