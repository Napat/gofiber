# heimdall: golang http client utilities

ปัญหาบน microservices

1. *One of the services may go down* Service ย่อยๆตัวใดตัวหนึ่งอาจจะล่ม (down) ลงไปได้ ส่งผลให้ services อื่นๆที่ขึ้นกับมันล่มตามลงไปด้วย
2. *One of the services may be taking too long* Service ใดๆอาจจะทำงานช้าหรือใช้เวลาในการทำงานนานขึ้น(slow down) จะทำให้ services อื่นๆที่ขึ้นกับมันทำงานช้าลงไปด้วย
3. *Growing pains* เมื่อมีการใช้งานเยอะๆ (heavy load) อาจจะส่งผลให้ services ช้าลง(slow down) หรือล่ม(down) ลงไปได้

แนวทางแก้ปัญหาที่ฝั่ง client ต้องเขียนตัวจัดการ(้handle failures) เพื่อช่วยหยุดปัญหาให้อยู่ที่เฉพาะ service ที่มีปัญหา(localized) ไม่ลุกลามไปยัง services อื่นๆหรือเกิดอาการ memory บวมขึ้นเรื่อยๆเนื่องจากหม่สามารถทำงานได้สำเร็จ

`Heimdall` คือ http client ที่มีฟีเจอร์สำหรับงาน microservices เพื่อช่วยลดความซับซ้อนในการเขียนโปรแกรมด้วย standard net/http พื้นฐานของ golang

Features

- `Retry mechanisms`: รองรับหลายๆ mechanisms เช่น Constant Backoff, Exponential Backoff และสามารถเขียน Custom retry mechanisms เองก็ได้
- `Plugins`: คล้ายการใช้ middleware บน golang http server แต่แบ่งย่อยเพิ่มออกมาเป็น 3 events คือ OnRequestStart(), OnRequestEnd(), OnError() ถ้าเทียบกับ middleware .Use() ของ server ก็คือมีแต่ event OnRequestStart()
- `Custom HTTP clients`: เป็นการ edit ความสามารถของ client ในเลเวลของ method .Do(request *http.Request) เช่นอยากให้เพิ่ม header บางอย่างไปในทุกๆ request ที่ใช้ client ที่สร้างขึ้นมา เป็นต้น
- `Circuit breaker`: เมื่อเกิด error เข้าตามเงื่อนไขสามารถจัดการเพิ่มได้เช่น register fallbackFunc ลงไป ....

## Retries

การ retry request เป็นกระบวนการที่ง่ายและมีประสิทธิภาพสูง การทำ retry มีเรื่องต้องพิจารณาดังนี้

*Service ดังกล่าวสามารถ retry ได้หรือไม่* เช่นถ้าเป็น Authentication หรือ Validation error ก็ไม่ควรเกิดการ retry ขึ้น
หรือหาก service ดังกล่าวสามารถ retry ได้ ก็ต้องพิจารณาต่อว่าจะ retry แค่บาง error code ใช่หรือไม่  
การทำงานของ `Heimdall` ใช้เงื่อนไขว่า ถ้า `response.StatusCode >= http.StatusInternalServerError` ถึงจะ retry ขยายความคือ ถ้าได้รับ http status code 401 ก็จะไม่เกิดการ retry แต่ถ้าได้ code 500 จะทำการ retry นั่นเอง หากเราต้องการเปลี่ยนแปลงเงื่อนไขนี้ต้องเขียน `Custom HTTP clients` เพื่อ implement การทำงานของ method .Do() ใหม่เอง

*ต้องพิจารณาว่าจะ retry สูงสุดกี่ครั้ง* การ Retry ใช้แก้ไขปัญหาในกรณีที่ Error เกิดแบบไม่ต่อเนื่อง(intermittent) หมายความว่า เราคาดหวังให้ไม่เกิด error ขึ้นในการ request รอบที่สอง หรือรอบที่สาม เพราะการ retry หลายๆครั้งอาจทำให้เกิดการส่ง requests ที่ยังตกค้างเพิ่มเป็นจำนวนมากได้(huge increase in the total requests being sent) ทำให้เกิดปัญหา high load เพิ่มขึ้นมาทั้งระบบโดยไม่จำเป็น(จากปัญหาแค่ service ติดๆดับๆแค่ service หนึ่ง)

*ระยะเวลาในการ retry แต่ละครั้งจะห่างเท่ากันไหร่* ต้องพิจารณาพฤติกรรมของ intermittent error ว่าเป็นอย่างไร
จะใช้การรอเป็นช่วงค่าคงที่ค่าหนึ่ง Constant Backoff  
จะใช้ Exponential Backoff เพื่อเพิ่มเวลาขึ้นเรื่อยๆในการ retry รอบถัดๆไป  
หรือจะต้องเขียน Custom retry mechanisms เองเนื่องจากเงื่อนไขของ neighbour/partner/business logic ที่มีความเฉพาะทาง

## Circuit breakers

เป็น concept เพื่อป้องกัน application เสียหายในกรณีที่มี high load ที่ service หนึ่งๆ  
เมื่อเข้าชุดของเงื่อนไขที่กำหนดเอาไว้ (violated restrictions) การส่งข้อมูลกันระหว่าง service จะหยุดลง (อาจจะหยุดแค่ระยะเวลาช่วงหนึ่งตามที่ตั้งค่าไว้)  
ชุดของข้อจำกัดขึ้นมาได้แก่ ปริมาณ/ความถี่ของ requests ที่เข้ามา, ตัวเลข consecutive errors ที่ค่าๆหนึ่งเป็น 10 errors ติดต่อกัน, ค่า Timeout: ระยะเวลาของ response time จาก downstream service เป็นต้น  

### Circuit breakers restrictions

`Timeout` การกำหนด Timeout เพื่อระบุเวลาในการรอ response จาก service

`Maximum concurrent requests` คือตัวเลข density สูงที่สุดที่ยังเหลืออยู่ในแต่ละช่วงของ services หากตรวจพบว่าค่ามากเกินก็จะตีความว่าเกิด error ขึ้น  
สาเหตุของปัญหานี้เกิดได้ 2 ลักษณะคือ  

1. downstream services ทำงานช้าลงทำให้ requests ที่ตกค้างเยอะขึ้น
2. ปริมาณ request สูงขึ้นจากทางฝั่งต้นทาง คือมีการใช้ service สูงขึ้น

การกำหนด(capping) concurrent requests สูงที่สุดที่ service รองรับได้เอาไว้เป็นแนวทางหนึ่งเพิ่อรักษาเสถียรภาพของระบบเอาไว้

### Error threshold percentage

เมื่อระบบมีปัญหาเกิดขึ้น(ตามเงื่อนไขของ restrictions ต่างๆ) จนกระทั่ง % error ของ request/respond สูงขึ้นเกินค่าที่กำหนดไว้ระบบก็จะหยุดการทำงานของ handler ไปยัง downstream service ลง
คือ เกิด circuit breaker trips ขึ้น เราเรียก state ที่หยุดการส่งต่อข้อมูลนี้ว่าสถานะ `Open ciecuit`

### Sleep window

แนวการจัดการหลังจากเกิด circuit breaker trips ขึ้นมีหลายวิธี  
วิธีการที่ hystrix ใช้คือการหยุดการส่ง requests ออกไปช่วงระยะเวลาหนึ่ง (`sleep window`) ตามที่เราตั้งค่า config เอาไว้  
หลังจากพ้นเวลาดังกล่าวแล้วค่อยส่ง request ออกไปอีกจำนวนหนึ่ง(`request volume threshold`) เพื่อทดสอบว่าปลายทางใช้งานได้แล้วหรือยัง  
หากใช้ได้แล้วก็กลับสู่สถานะ `Close circuit` ทำงานไปตามปกติ  

หน้าที่ของ `sleep window` คือเพื่อให้เวลา downstream service ได้ทำการ recovery ตัวเองได้ ดังนั้นค่าเวลาไม่ควรจะน้อยเกินไป  
และในทางตรงกันข้ามก็ไม่ควรมากเกินไปเช่นกันเพราะจะทำให้กระทบต่อการใช้งานที่ยังเหลืออยู่ด้วย(existing users)
ตัวอย่างเช่น เราอาจจะตั้งค่า sleep window ไว้ที่ `5 วินาที` อย่างไรก็ดีค่านี้ขึ้นอยู่กับธรรมชาติของ application นั้นๆด้วย

### request volume threshold

การทำงานของ Circuit breakers จำเป็นจะต้องได้รับ requests เข้ามาจำนวนหนึ่งก่อนถึงจะเอามาพิจารณาค่า Error threshold percentage ว่าวงจรควร open circuit เพื่อหยุดการส่ง requests ต่อไปยัง downstream  
จำนวน requests ขั้นต่ำดังกล่าวก็คือ `request volume threshold` ซึ่งเป็นค่าหนึ่งที่จำเป็นต้อง config เข้าไปในตอนเริ่มการทำงาน  
เราอาจจะเรียกช่วงที่จำนวน requests น้อยกว่าค่า `request volume threshold` นี้ว่า state `Half-open` ก็ได้

การทำงานของช่วง state `Open ciecuit` ระบบจะ reset ตัวนับต่างๆเช่น จำนวน requests ใหม่และหยุดส่งต่อ requests ไปช่วงระยะเวลาหนึ่ง(`sleep window`)  
จากนั้นระบบจะเข้าสู่ช่วง state `Half-open` เพื่อเริ่มส่ง requests ออกไปใหม่จนกระทั่งจำนวนของ requests ในรอบใหม่นี้มีจำนวนมากกว่าค่า `request volume threshold` จึงค่อยเข้าสู่ state `Closed circuit` เพื่อพิจารณาตัวเลข `Error threshold percentage` อีกครั้งหนึ่ง

## Heimdall / hystrix

ปกติการสร้าง client แบบง่ายๆเพื่อต้องการใช้แค่ feature ในการ retry เราสามารถใช้ `client := httpclient.NewClient( ... )` จาก `"github.com/gojektech/heimdall/v6/httpclient"` เพื่อสร้าง client ขึ้นมาก็เพียงพอ  

แต่เนื่องจาก Feature circuit breaker บน Heimdall เบื้องหลังจะเป็นการต่อยอดเอา [hystrix-go](https://github.com/afex/hystrix-go) มาใช้งาน  
ดังนั้นหากจะใช้งาน circuit breaker เราจะต้องสร้าง client ใน level ของ hystrix แทน(`"github.com/gojektech/heimdall/v6/hystrix"`)
ด้วยการเรียกคำสั่ง `client := hystrix.NewClient( ... )`  

## References

- [heimdall github](https://github.com/gojek/heimdall)
- [gojek blog: heimdall](https://www.gojek.io/blog/how-go-jek-handles-microservices-communication-at-scale)
- [hystrix and resilience: Golang](https://callistaenterprise.se/blogg/teknik/2017/09/11/go-blog-series-part11/)
- [netflix-hystrix: JAVA Spring](https://developers.ascendcorp.com/ทำ-microservices-ให้ยืดหยุ่นและแข็งแกร่งยิ่งกว่าเดิม-ด้วย-spring-cloud-และ-netflix-hystrix-af14ba952c46)
