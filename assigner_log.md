Старт сервиса:

```json
{"level":"INFO","ts":"2024-12-19T16:35:59.156Z","msg":"Cache Service configured successfuly"}
{"level":"INFO","ts":"2024-12-19T16:35:58.952Z","msg":"DB configured successfuly"}
{"level":"INFO","ts":"2024-12-19T16:35:58.952Z","msg":"HTTP Server configured successfuly"}
{"level":"INFO","ts":"2024-12-19T16:35:58.952Z","msg":"Starting HTTP Server..."}
```

curl -X POST 'http://localhost:5252/v1/assign_order?order-id=1&executor-id=1' -v

Добавляется:

```json
{"level":"INFO","ts":"2024-12-19T16:40:23.196Z","msg":"query started","request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"","method":""}
{"level":"INFO","ts":"2024-12-19T16:40:23.197Z","msg":"start building jobs graph","order_id":"1","executor_id":"1","request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:40:23.197Z","msg":"built jobs graph, start fetching","n_jobs":5,"request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:40:23.197Z","msg":"starting fetch from data source","data_source":"assign_order_configs","request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:40:23.197Z","msg":"starting fetch from data source","data_source":"executor_profile","request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:40:23.197Z","msg":"starting fetch from data source","data_source":"general_order_info","request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:40:23.198Z","msg":"fetched data source","data_source":"general_order_info","fetch_time":0.000939438,"request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:40:23.198Z","msg":"fetched data source","data_source":"executor_profile","fetch_time":0.000946062,"request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:40:23.198Z","msg":"starting fetch from data source","data_source":"zone_info","request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:40:23.198Z","msg":"fetched data source","data_source":"assign_order_configs","fetch_time":0.001078859,"request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:40:23.198Z","msg":"fetched data source","data_source":"zone_info","fetch_time":0.000238458,"request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:40:23.198Z","msg":"starting fetch from data source","data_source":"toll_roads_info","request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:40:23.198Z","msg":"fetched data source","data_source":"toll_roads_info","fetch_time":0.000153034,"request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:40:23.198Z","msg":"fetched all data sources","total_fetch_time":0.001497596,"request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:40:23.205Z","msg":"request to create order is processed","order_id":"1","executor_id":"1","request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"assigner","method":"assign_order"}
{"level":"INFO","ts":"2024-12-19T16:40:23.205Z","msg":"query finished","elapsed":0.008394574,"request_id":"ba0a023f8afe/hi9r1IUo7k-000001","service":"","method":""}
```

cancel:

```json
{"level":"INFO","ts":"2024-12-19T16:43:30.308Z","msg":"request to cancel order is processed","order_id":"1","request_id":"ba0a023f8afe/hi9r1IUo7k-000002","service":"assigner","method":"cancel_order"}
{"level":"INFO","ts":"2024-12-19T16:43:30.308Z","msg":"query finished","elapsed":0.021840803,"request_id":"ba0a023f8afe/hi9r1IUo7k-000002","service":"","method":""}
{"level":"INFO","ts":"2024-12-19T16:43:40.319Z","msg":"query started","request_id":"ba0a023f8afe/hi9r1IUo7k-000002","service":"","method":""}
```

assign второй раз. Пример, где для assign_order_configs просрочился cache. Происходит дополнительный поход в источники. Для zone_info данные подтягиваются из кэша.

```json
{"level":"INFO","ts":"2024-12-19T16:43:49.167Z","msg":"query started","request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"","method":""}
{"level":"INFO","ts":"2024-12-19T16:43:49.168Z","msg":"start building jobs graph","order_id":"1","executor_id":"1","request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:43:49.168Z","msg":"built jobs graph, start fetching","n_jobs":5,"request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:43:49.168Z","msg":"starting fetch from data source","data_source":"assign_order_configs","request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:43:49.168Z","msg":"starting fetch from data source","data_source":"general_order_info","request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:43:49.168Z","msg":"starting fetch from data source","data_source":"executor_profile","request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:43:49.168Z","msg":"fetched data source","data_source":"general_order_info","fetch_time":0.000726658,"request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:43:49.168Z","msg":"fetched data source","data_source":"assign_order_configs","fetch_time":0.000756572,"request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"cache_service","method":"get_order_info"}
```
Важный лог:
```json
{"level":"INFO","ts":"2024-12-19T16:43:49.168Z","msg":"retrieved data from cache","data_source":"zone_info","cache_key":"1","request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"cache_service","method":"get_order_info"}
```
```json
{"level":"INFO","ts":"2024-12-19T16:43:49.168Z","msg":"starting fetch from data source","data_source":"toll_roads_info","request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:43:49.168Z","msg":"fetched data source","data_source":"executor_profile","fetch_time":0.000754034,"request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:43:49.168Z","msg":"fetched data source","data_source":"toll_roads_info","fetch_time":0.000144412,"request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:43:49.169Z","msg":"fetched all data sources","total_fetch_time":0.000985953,"request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:43:49.175Z","msg":"request to create order is processed","order_id":"1","executor_id":"1","request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"assigner","method":"assign_order"}
{"level":"INFO","ts":"2024-12-19T16:43:49.175Z","msg":"query finished","elapsed":0.007509442,"request_id":"ba0a023f8afe/hi9r1IUo7k-000003","service":"","method":""}
```

Делаем три операции:
assign order_id = 2, executor_id = 2
cancel order_id = 2
assign order_id = 2, executor_id = 2

Тут все операции произошли быстро, поэтому для конфига тоже успело использоваться закешированное значение.

```json
{"level":"INFO","ts":"2024-12-19T16:47:21.857Z","msg":"query started","request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"","method":""}
{"level":"INFO","ts":"2024-12-19T16:47:21.857Z","msg":"start building jobs graph","order_id":"2","executor_id":"2","request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:21.857Z","msg":"built jobs graph, start fetching","n_jobs":5,"request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:21.857Z","msg":"starting fetch from data source","data_source":"assign_order_configs","request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:21.857Z","msg":"starting fetch from data source","data_source":"general_order_info","request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:21.857Z","msg":"starting fetch from data source","data_source":"executor_profile","request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:21.857Z","msg":"fetched data source","data_source":"executor_profile","fetch_time":0.000664059,"request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:21.857Z","msg":"fetched data source","data_source":"assign_order_configs","fetch_time":0.000705748,"request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:21.857Z","msg":"fetched data source","data_source":"general_order_info","fetch_time":0.000688864,"request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:21.857Z","msg":"starting fetch from data source","data_source":"zone_info","request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:21.858Z","msg":"fetched data source","data_source":"zone_info","fetch_time":0.000119975,"request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:21.858Z","msg":"starting fetch from data source","data_source":"toll_roads_info","request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:21.858Z","msg":"fetched data source","data_source":"toll_roads_info","fetch_time":0.000156865,"request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:21.858Z","msg":"fetched all data sources","total_fetch_time":0.001069462,"request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:21.864Z","msg":"request to create order is processed","order_id":"2","executor_id":"2","request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"assigner","method":"assign_order"}
{"level":"INFO","ts":"2024-12-19T16:47:21.864Z","msg":"query finished","elapsed":0.007590504,"request_id":"ba0a023f8afe/hi9r1IUo7k-000006","service":"","method":""}
{"level":"INFO","ts":"2024-12-19T16:47:28.555Z","msg":"query started","request_id":"ba0a023f8afe/hi9r1IUo7k-000007","service":"","method":""}
{"level":"INFO","ts":"2024-12-19T16:47:28.588Z","msg":"request to cancel order is processed","order_id":"2","request_id":"ba0a023f8afe/hi9r1IUo7k-000007","service":"assigner","method":"cancel_order"}
{"level":"INFO","ts":"2024-12-19T16:47:28.588Z","msg":"query finished","elapsed":0.032774352,"request_id":"ba0a023f8afe/hi9r1IUo7k-000007","service":"","method":""}
{"level":"INFO","ts":"2024-12-19T16:47:35.999Z","msg":"query started","request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"","method":""}
{"level":"INFO","ts":"2024-12-19T16:47:35.999Z","msg":"start building jobs graph","order_id":"2","executor_id":"2","request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:35.999Z","msg":"built jobs graph, start fetching","n_jobs":5,"request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"cache_service","method":"get_order_info"}
```
Важный лог:

```json
{"level":"INFO","ts":"2024-12-19T16:47:35.999Z","msg":"retrieved data from cache","data_source":"assign_order_configs","cache_key":"x","request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"cache_service","method":"get_order_info"}
```
```json
{"level":"INFO","ts":"2024-12-19T16:47:35.999Z","msg":"starting fetch from data source","data_source":"general_order_info","request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:35.999Z","msg":"starting fetch from data source","data_source":"executor_profile","request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:36.000Z","msg":"fetched data source","data_source":"general_order_info","fetch_time":0.000309646,"request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"cache_service","method":"get_order_info"}
```
Важный лог:
```json
{"level":"INFO","ts":"2024-12-19T16:47:36.000Z","msg":"retrieved data from cache","data_source":"zone_info","cache_key":"2","request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"cache_service","method":"get_order_info"}
```
```json
{"level":"INFO","ts":"2024-12-19T16:47:36.000Z","msg":"starting fetch from data source","data_source":"toll_roads_info","request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:36.000Z","msg":"fetched data source","data_source":"executor_profile","fetch_time":0.0003569,"request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:36.000Z","msg":"fetched data source","data_source":"toll_roads_info","fetch_time":0.000125685,"request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:36.000Z","msg":"fetched all data sources","total_fetch_time":0.000515814,"request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"cache_service","method":"get_order_info"}
{"level":"INFO","ts":"2024-12-19T16:47:36.006Z","msg":"request to create order is processed","order_id":"2","executor_id":"2","request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"assigner","method":"assign_order"}
{"level":"INFO","ts":"2024-12-19T16:47:36.006Z","msg":"query finished","elapsed":0.007044797,"request_id":"ba0a023f8afe/hi9r1IUo7k-000008","service":"","method":""}
```
