# Hava Kalitesi İzleme Sistemi

## İçindekiler
- [Projenin Amacı ve Kapsamı](#projenin-amacı-ve-kapsamı)
- [Sistem Mimarisi](#sistem-mimarisi)
  - [Backend Bileşenleri](#backend-bileşenleri)
  - [Veri Depolama](#veri-depolama)
  - [Mesaj Aracısı](#mesaj-aracısı)
  - [Ön Uç](#ön-uç)
  - [İletişim Akışları](#i̇letişim-akışları)
- [Teknoloji Seçimleri](#teknoloji-seçimleri)
  - [Go (Golang)](#go-golang)
  - [PostgreSQL/TimescaleDB](#postgresqltimescaledb)
  - [RabbitMQ](#rabbitmq)
  - [Next.js/React](#nextjsreact)
  - [Leaflet/React-Leaflet](#leafletreact-leaflet)
  - [Docker/Docker Compose](#dockerdocker-compose)
- [Kurulum ve Yapılandırma](#kurulum-ve-yapılandırma)
  - [Ön Koşullar](#ön-koşullar)
  - [Kurulum Adımları](#kurulum-adımları)
  - [Geliştirme Ortamı](#geliştirme-ortamı)
- [Kullanım Rehberi](#kullanım-rehberi)
  - [Web Arayüzü](#web-arayüzü)
  - [Veri Giriş Yöntemleri](#veri-giriş-yöntemleri)
- [API Dokümantasyonu](#api-dokümantasyonu)
  - [Veri Alım API](#veri-alım-api)
  - [Anomali API](#anomali-api)
  - [WebSocket API](#websocket-api)
- [Script Kullanımı](#script-kullanımı)
  - [manual-input.sh](#manual-inputsh)
  - [auto-test.sh](#auto-testsh)
- [Sorun Giderme](#sorun-giderme)
  - [Yaygın Sorunlar](#yaygın-sorunlar)
  - [Günlükler ve Hata Ayıklama](#günlükler-ve-hata-ayıklama)
  - [Sistemi Sıfırlama](#sistemi-sıfırlama)

Coğrafi görselleştirme yeteneklerine sahip gerçek zamanlı hava kalitesi izleme ve anomali tespit sistemi.

## Projenin Amacı ve Kapsamı

Bu proje, aşağıdaki amaçlara yönelik kapsamlı bir hava kalitesi izleme sistemi sunmaktadır:

1. Çeşitli kaynaklardan hava kalitesi ölçümlerini toplama
2. WHO standartları ve istatistiksel yöntemlere dayalı anomalileri tespit etme
3. Hava kalitesi verilerinin gerçek zamanlı uyarılarını ve görselleştirmelerini sağlama
4. Hava kalitesi desenlerinin coğrafi analizini destekleme

Sistem, özellikle yaygın hava kirleticilerinin (PM2.5, PM10, NO2, SO2, O3) coğrafi bir bölge genelinde olağandışı veya endişe verici değerlerini tespit etmeye, bu anomalileri interaktif haritalarda görüntülemeye ve karar vericiler ile çevre izleme uzmanları için analiz araçları sağlamaya odaklanmıştır.

## Sistem Mimarisi

Sistem, aşağıdaki temel bileşenlere sahip bir mikroservis mimarisini takip etmektedir:

### Backend Bileşenleri
1. **Veri Alım Servisi** (`air-quality-ingest`)
   - REST API aracılığıyla hava kalitesi ölçümlerini alır
   - İşlenmek üzere verileri mesaj kuyruğuna gönderir

2. **Ölçüm İşlemcisi** (`air-quality-processor`)
   - Kuyruktan ham ölçümleri alır
   - Anomali tespit algoritmalarını uygular
   - Normal ölçümleri TimescaleDB'de saklar
   - Anomalileri özel kuyruğa aktarır

3. **Anomali İşlemcisi** (`anomaly-processor`)
   - Anomali verilerini yönetir (depolama, geri alma)
   - Gerçek zamanlı anomali güncellemeleri için WebSocket sağlar
   - Anomali verilerini sorgulama için REST API sunar
   - Coğrafi yoğunluk hesaplamalarını destekler

### Veri Depolama
- **PostGIS uzantılı TimescaleDB**
  - Zaman serisi için optimize edilmiş veritabanı
  - Konum tabanlı sorgular için coğrafi yetenekler
  - Ölçümler ve anomaliler için ayrı tablolar

### Mesaj Aracısı
- **RabbitMQ**
  - Servisler arasında asenkron iletişimi yönetir
  - İki ana kuyruk: "measurements" ve "anomaly_alerts"

### Ön Uç
- **Next.js Web Uygulaması**
  - Anomali işaretleyicileriyle gerçek zamanlı harita görüntüleme
  - Anomali yoğunluğunun ısı haritası görselleştirmesi
  - Anomalilerin zamansal analizi için grafikler
  - Özel coğrafi sorgular için analiz araçları

### İletişim Akışları
1. **Veri Alım Akışı**
   - Harici kaynaklar → Veri Alım REST API → RabbitMQ → Ölçüm İşlemcisi → TimescaleDB

2. **Anomali Tespit Akışı**
   - Ölçüm İşlemcisi → Anomali Kuyruğu → Anomali İşlemcisi → TimescaleDB

3. **Veri Erişim Akışı**
   - Ön Uç → Anomali İşlemcisi REST API → TimescaleDB
   - Ön Uç ← WebSocket ← Anomali İşlemcisi (gerçek zamanlı güncellemeler)

## Teknoloji Seçimleri

### Go (Golang)
Backend servisleri için seçildi çünkü:
- Güçlü eşzamanlılık desteği (goroutine'ler, kanallar)
- Mükemmel performans özellikleri
- Daha iyi güvenilirlik için statik yazım
- Basit dağıtım (tek ikili dosya)

### PostgreSQL/TimescaleDB
Veri depolama için tercih edildi çünkü:
- TimescaleDB uzantısıyla zaman serisi optimizasyonu
- PostGIS entegrasyonu ile coğrafi yetenekler
- Güçlü sorgu yetenekleri ve ACID uyumluluğu
- Mükemmel dokümantasyona sahip olgun ekosistem

### RabbitMQ
Mesaj aracısı olarak uygulandı çünkü:
- Güvenilirlik ve kanıtlanmış üretim kullanımı
- Çoklu mesajlaşma paternlerine destek
- Servisler arasında net bir ayrım
- Mesaj kalıcılığı ve teslim garantileri

### Next.js/React
Ön uç teknoloji yığını seçildi çünkü:
- Sunucu taraflı render etme yetenekleri
- UI geliştirme için React bileşen modeli
- Güçlü TypeScript entegrasyonu
- Mükemmel geliştirici deneyimi ve performans

### Leaflet/React-Leaflet
Harita görselleştirme kütüphanesi seçildi çünkü:
- Kullanım kısıtlaması olmayan açık kaynak
- Geniş eklenti ekosistemi (ısı haritaları, işaretçiler)
- Hafif ve performanslı
- İyi belgelendirilmiş React entegrasyonu

### Docker/Docker Compose
Konteynerizasyon yaklaşımı seçildi çünkü:
- Tutarlı geliştirme ve üretim ortamları
- Çoklu servislerin basit orkestrayonu
- İzole servis bağımlılıkları
- Kolay yatay ölçekleme

## Kurulum ve Yapılandırma

### Ön Koşullar
- Docker ve Docker Compose
- Make (isteğe bağlı, script yürütmesi için)
- Web uygulaması için Chrome, Edge veya Opera gibi Chrome tabanlı bir tarayıcı (özellikle konum tabanlı özelliklerin en iyi şekilde çalışması için)

### Kurulum Adımları

1. **Depoyu klonlayın**
   ```bash
   git clone https://github.com/erentaskiran/kartacastaj.git air-quality-system
   cd air-quality-system
   ```

2. **Ortam Yapılandırması**
   
   Kök dizinde aşağıdaki değişkenleri içeren bir `.env` dosyası oluşturun (örnek değerler sağlanmıştır):
   ```
   # Veritabanı
   POSTGRES_USER=postgres
   POSTGRES_PASSWORD=password
   POSTGRES_DB=air_quality
   
   # RabbitMQ
   RABBITMQ_DEFAULT_USER=user
   RABBITMQ_DEFAULT_PASS=password
   ```

3. **Servisleri Başlatın**
   ```bash
   docker-compose up -d
   ```
   
   Bu, aşağıdakileri başlatacaktır:
   - TimescaleDB (PostGIS ile PostgreSQL)
   - RabbitMQ
   - Veri alım servisi (port 8000)
   - Ölçüm işlemcisi
   - Anomali işlemcisi (WebSocket port 8080, REST API port 8081)
   - Ön uç (port 3000)

4. **Servisleri Doğrulayın**
   
   Tüm servislerin çalıştığını kontrol edin:
   ```bash
   docker-compose ps
   ```

5. **Veritabanı Başlatma**
   
   Gerekli veritabanı tabloları TimescaleDB konteynerine bağlanan `db-setup.sql` scripti aracılığıyla otomatik olarak oluşturulacaktır.

### Geliştirme Ortamı

Geliştirme amaçlı olarak, bileşenleri bireysel olarak çalıştırabilirsiniz:

1. **Ön Uç Geliştirme**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   
   Ön uç http://localhost:3000 adresinde hot-reloading özelliğiyle kullanılabilir olacaktır.

2. **Backend Servis Geliştirme**
   
   Her servis bağımsız olarak çalıştırılabilir:
   ```bash
   cd air-quality-ingest
   go run cmd/server/main.go
   ```

   Servisleri bağımsız olarak çalıştırırken uygun ortam değişkenlerini (RABBITMQ_URL, DATABASE_URL) ayarladığınızdan emin olun.

## Kullanım Rehberi

### Web Arayüzü

Aşağıdaki özellikleri kullanmak için http://localhost:3000 adresindeki web arayüzüne erişin:

1. **Ana Sayfa** - Mevcut konumunuza göre yakındaki anomalileri görüntüleyin
   * **Not:** Konum tabanlı hizmetler için Chrome, Edge veya Opera gibi Chrome tabanlı bir tarayıcı kullanılması önerilir. Safari veya Firefox'ta konum izinleri ile ilgili sorunlar yaşanabilir. Konum izni sorunları yaşarsanız tarayıcı ayarlarından site için konum erişimine izin verdiğinizden emin olun.
2. **Isı Haritası** - Hava kalitesi anomalilerinin yoğunluğunu gösteren interaktif ısı haritası
3. **Grafikler** - Zaman içindeki anomali verilerinin görsel sunumu
4. **Uyarılar** - Detaylı son anomali uyarılarının listesi
5. **Analiz** - Anomali desenlerini incelemek için özel bölge analiz aracı

### Veri Giriş Yöntemleri

Sistem, hava kalitesi verilerini birden çok yöntemle alabilir:

1. **Manuel Giriş Scripti**
   
   Ölçümleri manuel olarak göndermek için sağlanan scripti kullanın:
   ```bash
   ./manual-input.sh <enlem> <boylam> <parametre> <değer>
   ```
   
   Örneğin:
   ```bash
   ./manual-input.sh 41.0082 28.9784 pm2.5 35.7
   ```

2. **Otomatik Test Üreteci**
   
   Kontrollü anomali oranlarına sahip örnek ölçümler oluşturmak için:
   ```bash
   ./auto-test.sh --duration=300 --rate=10 --anomaly-chance=0.2
   ```

3. **REST API**
   
   Ölçümleri programlı bir şekilde veri alım API'sini kullanarak gönderin:
   ```bash
   curl -X POST "http://localhost:8000/api/ingest" \
     -H "Content-Type: application/json" \
     -d '{
       "latitude": 41.0082,
       "longitude": 28.9784,
       "parameter": "pm2.5",
       "value": 35.7
     }'
   ```

## API Dokümantasyonu

### Veri Alım API

**POST /api/ingest**

Yeni bir hava kalitesi ölçümü gönderin.

İstek gövdesi:
```json
{
  "latitude": 41.0082,
  "longitude": 28.9784,
  "parameter": "pm2.5",
  "value": 35.7,
  "timestamp": "2025-01-15T14:30:00Z" // İsteğe bağlı, belirtilmezse mevcut zaman kullanılır
}
```

Parametreler:
- `latitude` (gerekli): WGS84 formatında ondalık enlem
- `longitude` (gerekli): WGS84 formatında ondalık boylam
- `parameter` (gerekli): "pm2.5", "pm10", "no2", "o3", "so2" değerlerinden biri
- `value` (gerekli): Sayısal ölçüm değeri
- `timestamp` (isteğe bağlı): ISO8601 zaman damgası, belirtilmezse mevcut zaman kullanılır

Yanıt: Başarılı durumda HTTP 200 OK

### Anomali API

**GET /api/anomalies/location**

Bir noktanın belirli bir yarıçapı içindeki anomalileri alın.

Sorgu parametreleri:
- `lat` (gerekli): Merkez enlem
- `lon` (gerekli): Merkez boylam
- `radius` (gerekli): Kilometre cinsinden arama yarıçapı

Yanıt:
```json
[
  {
    "parameter": "pm2.5",
    "value": 35.7,
    "time": "2025-01-15T14:30:00Z",
    "latitude": 41.0082,
    "longitude": 28.9784,
    "description": "Eşik Değeri"
  },
  ...
]
```

**GET /api/anomalies/timerange**

Belirli bir zaman aralığındaki anomalileri alın.

Başlıklar:
- `X-Start-Time` (gerekli): ISO8601 başlangıç zamanı
- `X-End-Time` (gerekli): ISO8601 bitiş zamanı

Yanıt: /api/anomalies/location ile aynı format

**GET /api/anomalies/density**

Coğrafi bir sınırlayıcı kutu içindeki anomali yoğunluğunu hesaplayın.

Sorgu parametreleri:
- `minLat` (gerekli): Sınırlayıcı kutunun minimum enlemi
- `minLon` (gerekli): Sınırlayıcı kutunun minimum boylamı
- `maxLat` (gerekli): Sınırlayıcı kutunun maksimum enlemi
- `maxLon` (gerekli): Sınırlayıcı kutunun maksimum boylamı

Yanıt:
```json
{
  "41.01_28.95": 5,
  "41.02_28.97": 3,
  ...
}
```
Burada anahtarlar "{enlem}_{boylam}" grid hücre tanımlayıcıları, değerler ise anomali sayılarıdır.

### WebSocket API

**WS /ws/live**

Anomali uyarılarının gerçek zamanlı akışı.

Bu WebSocket uç noktasına bağlanarak, tespit edildikleri anda gerçek zamanlı anomali uyarı JSON nesnelerini alın:

```json
{
  "parameter": "pm2.5",
  "value": 35.7,
  "time": "2025-01-15T14:30:00Z",
  "latitude": 41.0082,
  "longitude": 28.9784,
  "description": "Eşik Değeri"
}
```

## Script Kullanımı

### manual-input.sh

Manuel olarak tek bir hava kalitesi ölçümü gönderin.

```bash
./manual-input.sh <enlem> <boylam> <parametre> <değer>
```

Parametreler:
- `enlem`: Ondalık enlem (örn., 41.0082)
- `boylam`: Ondalık boylam (örn., 28.9784)
- `parametre`: "pm2.5", "pm10", "no2", "o3", "so2" değerlerinden biri
- `değer`: Sayısal ölçüm değeri

Örnek:
```bash
./manual-input.sh 41.0082 28.9784 pm2.5 35.7
```

### auto-test.sh

Kontrollü anomali oranlarıyla rastgele hava kalitesi ölçümleri oluşturun.

```bash
./auto-test.sh [seçenekler]
```

Seçenekler:
- `--duration=<saniye>`: Testin çalışma süresi (varsayılan: 60)
- `--rate=<istek/saniye>`: Saniyedeki ölçüm sayısı (varsayılan: 5)
- `--anomaly-chance=<0-1>`: Anomali oluşturma olasılığı (varsayılan: 0.15)
- `--center-lat=<enlem>`: Oluşturulan noktalar için merkez enlemi (varsayılan: 40.7)
- `--center-lon=<boylam>`: Oluşturulan noktalar için merkez boylamı (varsayılan: 30.0)
- `--radius=<km>`: Oluşturulan noktalar için yarıçap (varsayılan: 25)
- `--url=<url>`: Özel veri alım URL'si (varsayılan: http://localhost:8000/api/ingest)

Örnek:
```bash
./auto-test.sh --duration=300 --rate=10 --anomaly-chance=0.2 --center-lat=41.0 --center-lon=29.0
```

## Sorun Giderme

### Yaygın Sorunlar

1. **Servislerin başlamaması**
   
   Gerekli portların hâlihazırda kullanımda olup olmadığını kontrol edin:
   ```bash
   lsof -i :5432   # TimescaleDB
   lsof -i :5672   # RabbitMQ
   lsof -i :8000   # Veri Alım API
   lsof -i :8080   # WebSocket sunucu
   lsof -i :8081   # Anomali API
   lsof -i :3000   # Ön uç
   ```

2. **Veritabanı bağlantı hataları**
   
   TimescaleDB'nin çalıştığını ve erişilebilir olduğunu doğrulayın:
   ```bash
   docker-compose logs timescaledb
   ```
   
   Bağlantı parametrelerini kontrol edin:
   ```bash
   docker-compose exec timescaledb psql -U postgres -d air_quality -c "\l"
   ```

3. **RabbitMQ bağlantı sorunları**
   
   RabbitMQ durumunu kontrol edin:
   ```bash
   docker-compose logs rabbitmq
   ```
   
   Kuyrukların mevcut olduğunu doğrulayın:
   ```bash
   docker-compose exec rabbitmq rabbitmqctl list_queues
   ```

4. **Ön uç görselleştirmesi veri göstermiyor**
   
   - Tarayıcı konsolunda CORS veya bağlantı hataları olup olmadığını kontrol edin
   - ws://localhost:8080/ws/live WebSocket bağlantısını doğrulayın
   - Anomali API'nin http://localhost:8081/api adresinden erişilebilir olduğundan emin olun

5. **Hiç anomali tespit edilmiyor**
   
   - Anomali tespiti için WHO eşiklerinin üzerinde değerler gönderdiğinizden emin olun
   - Ölçüm işlemcisi günlüklerini kontrol edin:
     ```bash
     docker-compose logs mesurement-processor-service
     ```

### Günlükler ve Hata Ayıklama

Servis günlüklerine erişim:

```bash
# Tüm servis günlüklerini görüntüleyin
docker-compose logs

# Belirli servis günlüklerini görüntüleyin
docker-compose logs ingest-service
docker-compose logs mesurement-processor-service
docker-compose logs anomaly-processor-service
docker-compose logs frontend

# Günlükleri gerçek zamanlı olarak takip edin
docker-compose logs -f anomaly-processor-service
```

### Sistemi Sıfırlama

Sistemi ve tüm verileri tamamen sıfırlamak için:

```bash
# Tüm servisleri durdurun
docker-compose down

# Hacim verilerini kaldırın
docker-compose down -v

# Yeniden başlatın
docker-compose up -d
```

Bu, tüm kaydedilen ölçümleri ve anomalileri kaldırarak temiz bir başlangıç sağlayacaktır.