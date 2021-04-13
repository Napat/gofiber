# Gofiber Example

Gofiber คือ web framework ภาษา Golang มีจุดขายว่าออกแบบ API มาให้คล้ายกับการใช้งาน Express ของฝั่ง Node.js  
ซึ่งจะช่วยให้นักพัฒนาที่เคยใช้ Express สามารถมาใช้ Golang ที่มี performance สูงได้อย่างคุ้นเคยลดเวลาเรียนรู้ลง

เบื้องหลังของ go fiber ไม่ได้ใช้ standard `net/http` ของ golang แต่เลือกใช้ `fasthttp` ซึ่งเคลมตัวเองว่าเป็น http engine ที่เร็วที่สุดของ Golang(ในเวลาที่เขียนนี้)

``` bash
go mod init github.con/napat/gofiber
go run ./cmd/fiberserver/
```

``` bash
$ TOKEN=$(curl -s --data "user=john&pass=doe" http://localhost:3000/login | grep token | cut -d\" -f4); echo $TOKEN
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZXhwIjoxNjE4ODU5OTUwLCJuYW1lIjoiSm9obiBEb2UifQ.NgdvIFlLCYjBdXXRa8dFi4BomizjvA3-npqrlr5B1RJRZgQHb7CG24FudTpmcm8QlHzTosRQLdHBre_mi1ak5zn2EtnYDAxzOcU_EeOJpO7-7041PJSRf2vEiV-Mq6x5mKv_aKYjS7no9rPcwjWrUquMaZCaenEnciWzESjf53lexHCstPrZwB4EQG2OcfznuyxBjU8aNm1N2BkS_hyTcpQ-4qsKbirl-DQWO_2nVChnuEZBXQ41nBnkW223jGSa4qAX6G_rZlEHbRkuq4ddHFvPsd9jQY6wqLI0Dw6KwCo3LBfH73opVjGfz3NvsBcPpeI9AcVz2uMQjPfAt2gbXw

$ curl -v localhost:3000/private -H "Authorization: Bearer $TOKEN"
...
Welcome John Doe
```

## Middleware

Gofiber มี [Internal middleware](https://github.com/gofiber/fiber#-internal-middleware) และ [External middleware](https://github.com/gofiber/fiber#-external-middleware) ที่พัฒนาโดย fiber team ให้ใช้งานอยู่จำนวนหนึ่งทำให้สามารถเริ่มใช้งานฟีเจอร์หลักๆได้อย่างรวดเร็ว  
นอกจากนี้ยังมี [Third Party Middlewares](https://github.com/gofiber/fiber#-third-party-middlewares) สำหรับทำงานร่วมกับโปรแกรมอื่นๆเช่น swagger, prometheus, tracing ด้วยเช่นกัน

## [Validation](https://docs.gofiber.io/guide/validation)

Guideline ของ gofiber แนะนำให้ไปใช้ [validator link](https://github.com/go-playground/validator) สำหรับการตรวจสอบข้อมูล struct ซึ่งจริงๆก็ไม่ได้ทำงานผูกกับตัว gofiber เลย  
หลักการทำงานคือ หลักจากรับ request เข้ามาที่ fiberHttpHandler แล้วก็เรียก c.BodyParser() เพื่อ parse ข้อมูลใส่ struct
แล้วจึงโยนเข้าไปใน function ที่เราเขียนขึ้นเพื่อ ValidateStruct ภายในก็เรียกใช้งาน package validator (validate.Struct()) เพื่อตรวจสอบ
แล้วจึงตอบว่ามี error หรือไม่กลับออกมายัง fiberHttpHandler  

## Test

Unit test

gofiber มี functions สำหรับช่วยทำ unit test ให้ใช้งานแล้วเช่น `app.Test(req)` ดูตัวอย่างได้ที่  

- [Link1](https://docs.gofiber.io/api/app#test)
- [Link2](https://github.com/gofiber/fiber/issues/41#issuecomment-768954313)
- [Link3](https://github.com/gofiber/fiber/blob/master/ctx_test.go)
- [Link4](https://github.com/gofiber/fiber/blob/master/middleware/cache/cache_test.go)

``` bash
go test -cover ./...
?       github.con/napat/gofiber/cmd/fiberserver          [no test files]
?       github.con/napat/gofiber/cmd/fiberserver_onefile  [no test files]
ok      github.con/napat/gofiber/internal/fiberserver     0.581s  coverage: 50.8% of statements

$ go test -p 2 ./internal/fiberserver/
ok      github.con/napat/gofiber/internal/fiberserver   0.674s

$ go test github.con/napat/gofiber/internal/fiberserver/
ok      github.con/napat/gofiber/internal/fiberserver   (cached)

$ go test -run TestPrivateHandler github.con/napat/gofiber/internal/fiberserver/
ok      github.con/napat/gofiber/internal/fiberserver   0.240s
```

Example load test

``` bash
go run ./cmd/fiberserver/
```

``` bash
$ wrk -t4 -c400 -d10s http://localhost:3000/Running 10s test @ http://localhost:3000/
  4 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   243.60ms  191.25ms   1.15s    85.13%
    Req/Sec   572.62    331.63     1.03k    50.70%
  17816 requests in 10.06s, 2.21MB read
  Socket errors: connect 0, read 14, write 0, timeout 0
Requests/sec:   1770.56
Transfer/sec:    224.78KB
```

## References

- [gofiber: route & group](https://github1s.com/gofiber/recipes/blob/master/auth-jwt/router/router.go)
- [gofiber template](https://github.com/gofiber/template)
- [Internal middleware](https://github.com/gofiber/fiber#-internal-middleware)
- [Internal middleware: Logger](https://github.com/gofiber/fiber/tree/master/middleware/logger)
- [Internal middleware: Cache](https://github.com/gofiber/fiber/tree/master/middleware/cache)
- [External middleware: JWT](https://github.com/gofiber/jwt)
- [Validation](https://docs.gofiber.io/guide/validation)
- [validator package](https://github.com/go-playground/validator)
- [Graceful shutdown](https://github.com/gofiber/recipes/tree/master/graceful-shutdown)
- [Golang TDD: Learn Go with Tests](https://quii.gitbook.io/learn-go-with-tests/)
- [Golang Test mock httt fail example](https://medium.com/@thejasbabu/testing-in-golang-c378b351002d)
- [fiber Test](https://docs.gofiber.io/api/app#test)
- [fiber Test](https://github.com/gofiber/fiber/issues/41#issuecomment-768954313)
- [fiber Test](https://github.com/gofiber/fiber/blob/master/ctx_test.go)
- [fiber Test](https://github.com/gofiber/fiber/blob/master/middleware/cache/cache_test.go)
- [project layout](https://dev.to/lucasnevespereira/write-a-rest-api-in-golang-following-best-practices-pe9)
- [giber example: CRUD](https://medium.com/cnxdevsoft/มาทำ-api-กับ-go-fiber-crudง่ายๆกันดีกว่า-459c7e8af548)
