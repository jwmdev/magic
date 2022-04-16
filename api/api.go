package api

/*
//func API(app *echo.Echo) {

//apiv1 := app.Group("/api/v1")
//apiv1.GET("/nodes/:hash/proxy-load.json",
func ProxyLoad(c echo.Context) error {

	hash := c.Param("hash")

	rows, _ := connectClickhouse.Query(
		`SELECT * FROM(
						SELECT
							timestamp,
							transaction_hash,
							from_address,
							to_address,
							visitParamExtractInt(data_string,'rps') rps,
							visitParamExtractInt(data_string,'qps') qps,
							visitParamExtractString( data_string,'mhaddr') mhaddr
						FROM Transactions
						WHERE status_int = 4353 AND mhaddr=?
						ORDER BY timestamp DESC
					) ORDER BY timestamp ASC`,
		hash)

	defer rows.Close()

	pointsRPS := [][]string{}
	pointsQPS := [][]string{}
	for rows.Next() {
		var (
			timestamp                                   time.Time
			rps, qps                                    int64
			transaction, fromAddress, toAddress, mhaddr string
		)
		if err := rows.Scan(&timestamp, &transaction, &fromAddress, &toAddress, &rps, &qps, &mhaddr); err != nil {
			log.Fatal(err)
		}

		nodeName := hashtrim(fromAddress)

		pointsRPS = append(pointsRPS, []string{timestamp.Format("2006-01-02 15:04"), nodeName, strconv.FormatInt(rps, 10)})
		pointsQPS = append(pointsQPS, []string{timestamp.Format("2006-01-02 15:04"), nodeName, strconv.FormatInt(qps, 10)})
	}

	result := struct {
		RPS [][]string `json:"rps"`
		QPS [][]string `json:"qps"`
	}{
		RPS: pointsRPS,
		QPS: pointsQPS,
	}

	return c.JSON(http.StatusOK, result)
}

//apiv1.GET("/nodes/:hash/delegations.json",

func Delegations(c echo.Context) error {
	hash := c.Param("hash")
	rows, _ := connectClickhouse.Query(`
		SELECT
			toStartOfInterval(timestamp, INTERVAL 1 hour) tt,
			countIf(method='delegate') delegate,
			countIf(method='undelegate') undelegate
		FROM Transactions
		WHERE method IN('delegate','undelegate') AND to_address=?
		GROUP BY tt
		ORDER BY tt ASC`, hash)
	defer rows.Close()

	times := []string{}
	delegates := []int64{}
	undelegates := []int64{}

	for rows.Next() {
		var (
			timestamp                      time.Time
			delegateCount, undelegateCount int64
		)
		if err := rows.Scan(&timestamp, &delegateCount, &undelegateCount); err != nil {
			log.Fatal(err)
			continue
		}

		times = append(times, timestamp.Format("Jan _2 15:04:05"))
		delegates = append(delegates, delegateCount)
		undelegates = append(undelegates, undelegateCount)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"time":        times,
		"delegates":   delegates,
		"undelegates": undelegates,
	})
}

//apiv1.GET("/nodes/:hash/rewards.json",

func Reward(c echo.Context) error {
	hash := c.Param("hash")
	rows, _ := connectClickhouse.Query(`
		SELECT
			date,
			sum(value/1e6)
		FROM Transactions
		WHERE
			status_int=102 AND to_address=?
		GROUP BY date
		ORDER BY date ASC`, hash)
	defer rows.Close()

	times := []string{}
	nodeRewards := []float64{}

	for rows.Next() {
		var (
			timestamp  time.Time
			forgingSum float64
		)
		if err := rows.Scan(&timestamp, &forgingSum); err != nil {
			log.Fatal(err)
			continue
		}

		nodeRewards = append(nodeRewards, forgingSum)
		times = append(times, timestamp.Format("Jan _2"))
	}

	return c.JSON(http.StatusOK, echo.Map{
		"time":    times,
		"rewards": nodeRewards,
	})
}

//apiv1.GET("/nodes.json",

func Nodes(c echo.Context) error {

	type point struct {
		Name            string  `json:"name"  db:"name"`
		Address         string  `json:"address"  db:"address"`
		CountryLong     string  `json:"country_long"  db:"country_long"`
		DelegatedAmount float64 `json:"delegated_amount"  db:"delegated_amount"`
	}

	nodes := []point{}
	err = connectMysql.Select(&nodes, `SELECT nodes.address, nodes.name, nodes.country_long, addresses.delegated_amount AS delegated_amount
			FROM nodes
			INNER JOIN addresses ON (nodes.address=addresses.address)
			ORDER BY nodes.last_updated DESC`)
	if err != nil {
		log.Fatal(err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": nodes,
	})
}

//apiv1.GET("/status.json",
func Status(c echo.Context) error {
	return c.JSON(http.StatusOK, getUpdateSystemStatus())
}

//apiv1.GET("/status/txs.json",
func Tx(c echo.Context) error {

	key := []byte("status_trx")
	valuesRaw, _ := cache.Get(key)

	if valuesRaw == nil {

		rows, _ := connectClickhouse.Query(`
				SELECT
					timestamp tt,
					count()
				FROM Transactions
				WHERE date >= today()-1 AND timestamp >= (NOW()-INTERVAL 24 HOUR)
				GROUP BY tt
				ORDER BY tt ASC`)
		defer rows.Close()

		times := []string{}
		trxCount := []int64{}
		for rows.Next() {
			var (
				timestamp time.Time
				count     int64
			)
			if err := rows.Scan(&timestamp, &count); err != nil {
				log.Fatal(err)
				continue
			}

			times = append(times, timestamp.Format("Jan _2 15:04:05"))
			trxCount = append(trxCount, count)
		}

		values := echo.Map{
			"time": times,
			"trx":  trxCount,
		}
		valuesMarshal := gotiny.Marshal(&values)
		_ = cache.Set(key, valuesMarshal, 5)
		return c.JSON(http.StatusOK, values)
	}

	valuesCache := echo.Map{}
	gotiny.Unmarshal(valuesRaw, &valuesCache)

	return c.JSON(http.StatusOK, valuesCache)
}

//apiv1.GET("/status/txs_date.json",
func TxDate(c echo.Context) error {

	key := []byte("txs_date")
	valuesRaw, _ := cache.Get(key)

	if valuesRaw == nil {

		rows, _ := connectClickhouse.Query(`
				SELECT
					date tt,
					count()
				FROM Transactions
				GROUP BY tt
				ORDER BY tt ASC`)
		defer rows.Close()

		times := []string{}
		trxCount := []int64{}
		for rows.Next() {
			var (
				timestamp time.Time
				count     int64
			)
			if err := rows.Scan(&timestamp, &count); err != nil {
				log.Fatal(err)
				continue
			}

			times = append(times, timestamp.Format("Jan _2"))
			trxCount = append(trxCount, count)
		}

		values := echo.Map{
			"time": times,
			"trx":  trxCount,
		}
		valuesMarshal := gotiny.Marshal(&values)
		_ = cache.Set(key, valuesMarshal, 5)
		return c.JSON(http.StatusOK, values)
	}

	valuesCache := echo.Map{}
	gotiny.Unmarshal(valuesRaw, &valuesCache)

	return c.JSON(http.StatusOK, valuesCache)
}

//apiv1.GET("/status/wallets.json",
func Wallets(c echo.Context) error {

	key := []byte("status_wallets")
	valuesRaw, _ := cache.Get(key)

	if valuesRaw == nil {

		rows, _ := connectClickhouse.Query(`
			SELECT
				date,
				countDistinct(to_address),
				countDistinct(from_address)
			FROM Transactions
			GROUP BY date
			ORDER BY date ASC`)
		defer rows.Close()

		times := []string{}
		walletsUniq := []int64{}
		walletsTotal := []int64{}
		for rows.Next() {
			var (
				timestamp   time.Time
				uniq, total int64
			)
			if err := rows.Scan(&timestamp, &total, &uniq); err != nil {
				log.Fatal(err)
				continue
			}

			times = append(times, timestamp.Format("Jan _2"))
			walletsTotal = append(walletsTotal, total)
			walletsUniq = append(walletsUniq, uniq)
		}

		values := echo.Map{
			"time":          times,
			"wallets_uniq":  walletsUniq,
			"wallets_total": walletsTotal,
		}
		valuesMarshal := gotiny.Marshal(&values)
		_ = cache.Set(key, valuesMarshal, 60*15)
		return c.JSON(http.StatusOK, values)
	}

	valuesCache := echo.Map{}
	gotiny.Unmarshal(valuesRaw, &valuesCache)

	return c.JSON(http.StatusOK, valuesCache)
}

//apiv1.GET("/status/delegations.json",
func Delegations(c echo.Context) error {

	key := []byte("status_delegations")
	valuesRaw, _ := cache.Get(key)

	if valuesRaw == nil {

		rows, _ := connectClickhouse.Query(`
			SELECT
				toStartOfInterval(timestamp, INTERVAL 1 hour) tt,
				countIf(method='delegate') delegate,
				countIf(method='undelegate') undelegate
			FROM Transactions
			WHERE method IN('delegate','undelegate') AND status_int=20 AND to_address<>'0x666174686572206f662077616c6c65747320666f7267696e67'
			GROUP BY tt
			ORDER BY tt ASC`)
		defer rows.Close()

		times := []string{}
		delegates := []int64{}
		undelegates := []int64{}

		for rows.Next() {
			var (
				timestamp                      time.Time
				delegateCount, undelegateCount int64
			)
			if err := rows.Scan(&timestamp, &delegateCount, &undelegateCount); err != nil {
				log.Fatal(err)
				continue
			}

			times = append(times, timestamp.Format("Jan _2 15:04:05"))
			delegates = append(delegates, delegateCount)
			undelegates = append(undelegates, undelegateCount)
		}

		values := echo.Map{
			"time":        times,
			"delegates":   delegates,
			"undelegates": undelegates,
		}
		valuesMarshal := gotiny.Marshal(&values)
		_ = cache.Set(key, valuesMarshal, 60*15)
		return c.JSON(http.StatusOK, values)
	}

	valuesCache := echo.Map{}
	gotiny.Unmarshal(valuesRaw, &valuesCache)

	return c.JSON(http.StatusOK, valuesCache)
}

//apiv1.GET("/status/delegation_sum.json",

func DelegationSum(c echo.Context) error {

	key := []byte("status_delegation_sum")
	valuesRaw, _ := cache.Get(key)

	if valuesRaw == nil {

		rows, _ := connectClickhouse.Query(`
				SELECT
					date,
					sum(delegate)/1e6 delegateSum
				FROM Transactions
				WHERE method='delegate' AND status_int=20 AND to_address<>'0x666174686572206f662077616c6c65747320666f7267696e67'
				GROUP BY date
				ORDER BY date ASC`)
		defer rows.Close()

		times := []string{}
		delegates := []float64{}

		for rows.Next() {
			var (
				timestamp   time.Time
				delegateSum float64
			)
			if err := rows.Scan(&timestamp, &delegateSum); err != nil {
				log.Fatal(err)
				continue
			}

			times = append(times, timestamp.Format("Jan _2"))
			delegates = append(delegates, delegateSum)
		}

		values := echo.Map{
			"time":      times,
			"delegates": delegates,
		}
		valuesMarshal := gotiny.Marshal(&values)
		_ = cache.Set(key, valuesMarshal, 60*15)
		return c.JSON(http.StatusOK, values)
	}

	valuesCache := echo.Map{}
	gotiny.Unmarshal(valuesRaw, &valuesCache)

	return c.JSON(http.StatusOK, valuesCache)
}

//apiv1.GET("/status/amount_sum.json",
func SumAmount(c echo.Context) error {

	key := []byte("status_amount_sum")
	valuesRaw, _ := cache.Get(key)

	if valuesRaw == nil {

		rows, _ := connectClickhouse.Query(`
				SELECT
					date tt,
					sum( value )/1e6 value
				FROM Transactions
				WHERE status_int=20 AND type_tx='block'
				GROUP BY tt
				ORDER BY tt ASC`)
		defer rows.Close()

		times := []string{}
		sum := []float64{}

		for rows.Next() {
			var (
				timestamp   time.Time
				delegateSum float64
			)
			if err := rows.Scan(&timestamp, &delegateSum); err != nil {
				log.Fatal(err)
				continue
			}

			times = append(times, timestamp.Format("Jan _2"))
			sum = append(sum, delegateSum)
		}

		values := echo.Map{
			"time": times,
			"sum":  sum,
		}
		valuesMarshal := gotiny.Marshal(&values)
		_ = cache.Set(key, valuesMarshal, 60*15)
		return c.JSON(http.StatusOK, values)
	}

	valuesCache := echo.Map{}
	gotiny.Unmarshal(valuesRaw, &valuesCache)

	return c.JSON(http.StatusOK, valuesCache)
}

//apiv1.GET("/status/forging.json",
func Forging(c echo.Context) error {

	key := []byte("status_forging")
	valuesRaw, _ := cache.Get(key)

	if valuesRaw == nil {

		rows, _ := connectClickhouse.Query(`
				SELECT
					toStartOfHour(timestamp) tt,
					countIf(method='start forging') starts,
					countIf(method='stop forging') stops
				FROM Transactions
				WHERE status_int=20 AND to_address='0x666174686572206f662077616c6c65747320666f7267696e67'
				GROUP BY tt
				ORDER BY tt ASC`)
		defer rows.Close()

		times := []string{}
		starts := []int64{}
		stops := []int64{}

		for rows.Next() {
			var (
				timestamp               time.Time
				startsCount, stopsCount int64
			)
			if err := rows.Scan(&timestamp, &startsCount, &stopsCount); err != nil {
				log.Fatal(err)
				continue
			}

			times = append(times, timestamp.Format("Jan _2 15:04"))
			starts = append(starts, startsCount)
			stops = append(stops, stopsCount)
		}

		values := echo.Map{
			"time":   times,
			"starts": starts,
			"stops":  stops,
		}
		valuesMarshal := gotiny.Marshal(&values)
		_ = cache.Set(key, valuesMarshal, 60*15)
		return c.JSON(http.StatusOK, values)
	}

	valuesCache := echo.Map{}
	gotiny.Unmarshal(valuesRaw, &valuesCache)

	return c.JSON(http.StatusOK, valuesCache)
}

//api for stakingrewards.com
//apiv1.GET("/status/sr.json",

func StakingReward(c echo.Context) error {

	//key := []byte("status_sr")

	type validatorDetail struct {
		Address           string  `json:"validator_address" db:"address"`
		ValidatorName     string  `json:"validator_name" db:"name"`
		ValidatorType     string  `json:"validator_type" db:"node_type"`
		ValidatorLocation string  `json:"validator_location" db:"location"`
		Delegated         float64 `json:"delegated_amount" db:"delegated"`
		Roi               float64 `json:"node_reward" db:"mg_roi"`
		DelegationCount   int     `json:"num_delegations" db:"is_online"` //This is just to keep the value zero
	}

	type sr struct {
		NodeReward        float64           `json:"validator_reward"`
		CoinReward        float64           `json:"coin_reward"`
		TotalReward       float64           `json:"total_reward"`
		Validators        int64             `json:"validators"`
		BlockTime         float64           `json:"block_time"`
		TransactionFee    float64           `json:"transaction_fee"`
		ActiveDelegators  int64             `json:"active_delegators"`
		TotalSupply       float64           `json:"total_supply"`
		TotalDelegated    float64           `json:"total_delegated"`
		CirculatingSupply float64           `json:"circulating_supply"`
		ValidatorsInfo    []validatorDetail `json:"validators_info"`
	}

	srData := &sr{NodeReward: 0.0, CoinReward: 0.0, TotalReward: 0.0, Validators: 0, BlockTime: 0.0, TransactionFee: 0.0, ActiveDelegators: 0}

	//get validators
	var validators int64
	sqlValidators := `select count(*) from nodes  where node_type='Core' or node_type='Verifier'`
	err := connectMysql.Get(&validators, sqlValidators)
	if err == nil {
		srData.Validators = validators
	}

	//get validators details
	validatorInfo := []validatorDetail{}
	sqlValDetails := `select nodes.address,  nodes.name, nodes.node_type, nodes.mg_geo as location, nodes.mg_roi,  (convert(addresses.delegated_amount,double)/1e6) as delegated, nodes.is_online from nodes, addresses where nodes.address=addresses.address and nodes.node_type IN ('Verifier','Core')`
	err = connectMysql.Select(&validatorInfo, sqlValDetails)
	if err == nil {
		srData.ValidatorsInfo = validatorInfo

		//count number of delegators
		for i, v := range srData.ValidatorsInfo {
			data := getLastNodeTrust(v.Address)
			d := 0
			if data != nil {
				c := data.CountNodeDelegations()
				d = c
			}
			srData.ValidatorsInfo[i].DelegationCount = d
			//time.Sleep(time.Millisecond * 20)
		}

	}
	// get time =  number of blocks in the past 24 hrs from now
	var blocks int64 = 0
	sqlBlocks := `SELECT countDistinct(block_number)  FROM Transactions WHERE timestamp>=(NOW()-INTERVAL 24 HOUR)`
	err = connectClickhouse.Get(&blocks, sqlBlocks)
	if err == nil {
		srData.BlockTime = 24.0 * 60.0 * 60.0 / float64(blocks) //calcualte block time 2hrs/block (number of seconds between blocks)
	}

	//participation rate select  count(distinct(address)) from addresses where frozen>512000000
	var participation int64 = 0
	//sqlActiveDelegators := `select count(DISTINCT(address)) from addresses where delegated>0`
	sqlActiveDelegators := `select  count(distinct(address)) from addresses where frozen>512000000`

	err = connectMysql.Get(&participation, sqlActiveDelegators)
	if err == nil {
		srData.ActiveDelegators = participation
	}

	// average node reward
	var nodeReward float64 = 0.0
	sqlNodeReward := `select avg(roi) from (select convert(max(mg_roi), double) as roi, mg_geo, node_type, count(node_type) as num_nodes from nodes where mg_geo<>'' and is_online=1 and node_type<>'Core'  group by mg_geo, node_type order by mg_geo asc, node_type asc) as tt;`
	err = connectMysql.Get(&nodeReward, sqlNodeReward)
	if err == nil {
		srData.NodeReward = nodeReward
	}
	//coin reward
	addr := "0x00e25887bcfd082c15e959100006cbbc006c1e8059ba43ed84"
	roi := calculateCoinRoi(addr)

	srData.CoinReward = roi

	// node reward
	srData.TotalReward = srData.NodeReward + srData.CoinReward

	srData.TotalSupply = 9200000000.0

	//get balance
	bal, _ := getCommonBalance()

	//get total delegated
	//dele := getAllNodesDelegation()

	var dele int64 = 0
	var dele2 float64 = 0
	sqlDelSum := `select sum(delegated_amount) from addresses`
	connectMysql.Get(&dele, sqlDelSum)
	if err == nil {
		dele2 = float64(dele) / 1e6
	}

	if bal > 0.0 && dele2 > 0.0 && bal > dele2 {
		srData.CirculatingSupply = bal - dele2
		srData.TotalDelegated = dele2
	}

	//TODO: Fix cache issue, data cannot be cached
	//valuesRaw, err := cache.Get(key)
	//if err != nil {
	//	log.Printf("Error getting cache %v\n", err)

	//}

	//if valuesRaw == nil {

	//	valuesMarshal := gotiny.Marshal(&srData)
	//	err := cache.Set(key, valuesMarshal, 1800)
	//	log.Printf("Error setting cache: %v\n", err)
	//	return c.JSON(http.StatusOK, srData)
	//}

	//valuesCache := echo.Map{}
	//gotiny.Unmarshal(valuesRaw, &valuesCache)

	return c.JSON(http.StatusOK, srData)
}

//apiv1.GET("/address/:address/txs.json",

func AddressTx(c echo.Context) error {

	address := c.Param("address")

	key := []byte("address_txs_" + address)
	valuesRaw, _ := cache.Get(key)

	var result struct {
		To       []float32 `json:"to" db:"toAcount"`
		From     []float32 `json:"from" db:"fromAcount"`
		ToSum    []float32 `json:"to_sum" db:"toAcountSum"`
		FromSumm []float32 `json:"from_sum" db:"fromAcountSum"`
		Time     []string  `json:"time" db:"tt"`
	}

	if valuesRaw == nil {

		sql := `SELECT
					toStartOfHour(timestamp) tt,
					sumIf(value,to_address=?)/1e6 toAcountSum,
					sumIf(value, from_address=?)/1e6 fromAcountSum,
					countIf(to_address=?) toAcount,
					countIf(from_address=?) fromAcount
				FROM Transactions
				WHERE status_int=20 AND (to_address=? OR from_address=?)
				GROUP BY tt
				ORDER BY tt ASC`

		type point struct {
			To       float32   `json:"to" db:"toAcount"`
			From     float32   `json:"from" db:"fromAcount"`
			ToSum    float32   `json:"to_sum" db:"toAcountSum"`
			FromSumm float32   `json:"from_sum" db:"fromAcountSum"`
			Time     time.Time `json:"time" db:"tt"`
		}

		points := []point{}

		err = connectClickhouse.Select(&points, sql, address, address, address, address, address, address)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		for _, point := range points {
			result.To = append(result.To, point.To)
			result.From = append(result.From, point.From)
			result.ToSum = append(result.ToSum, point.ToSum)
			result.FromSumm = append(result.FromSumm, point.FromSumm)
			result.Time = append(result.Time, point.Time.Format("2006-01-02 15:00"))
		}

		valuesMarshal := gotiny.Marshal(&result)
		_ = cache.Set(key, valuesMarshal, 60*15)
		return c.JSON(http.StatusOK, &result)
	}

	valuesCache := result
	gotiny.Unmarshal(valuesRaw, &valuesCache)

	return c.JSON(http.StatusOK, valuesCache)
}

//apiv1.GET("/address/:address/txs_stat.json",

func AddressStatistics(c echo.Context) error {

	address := c.Param("address")

	all := c.QueryParam("all")
	countTxs := c.QueryParam("countTxs")

	countTxsI, _ := strconv.Atoi(countTxs)
	if countTxsI == 0 && all == "" {
		countTxsI = txLimit
	}

	responseHistory, err := rpcClientTorrent.Call("fetch-history", &metawatch.HistoryArgs{Address: address, CountTxs: int64(countTxsI)})

	if err == nil {
		return c.JSON(http.StatusOK, &responseHistory)
	}

	return c.JSON(http.StatusBadRequest, err.Error())
}

//apiv1.GET("/status/size.json",
func BlockSize(c echo.Context) error {

	key := []byte("api_sizes")
	valuesRaw, _ := cache.Get(key)

	if valuesRaw == nil {

		type blockSizeModel struct {
			BlockHour string `json:"block_hour" db:"block_hour"`
			FullSize  int64  `json:"full_size" db:"full_size"`
		}

		var (
			listBlockSizes = []blockSizeModel{}
			sqlBlockSizes  = `SELECT block_hour, MAX(full_size) full_size FROM(
						SELECT DATE_FORMAT(timestamp,'%Y-%m-%d') as block_hour, SUM(size) over(order by number range between unbounded preceding and current row) full_size FROM blocks
					) s1
					GROUP BY block_hour`
		)

		err = connectMysql.Select(&listBlockSizes, sqlBlockSizes)
		if err != nil {
			log.Println(err.Error())
		}

		var (
			blockHour []string
			fullSize  []int64
		)

		for _, blockSizeInfo := range listBlockSizes {
			blockHour = append(blockHour, blockSizeInfo.BlockHour)
			fullSize = append(fullSize, blockSizeInfo.FullSize)
		}

		values := echo.Map{
			"block_hour": blockHour,
			"full_size":  fullSize,
		}
		valuesMarshal := gotiny.Marshal(&values)
		_ = cache.Set(key, valuesMarshal, 60*55)
		return c.JSON(http.StatusOK, values)
	}

	valuesCache := echo.Map{}
	gotiny.Unmarshal(valuesRaw, &valuesCache)

	return c.JSON(http.StatusOK, valuesCache)
}

//apiv1.GET("/status/blocks.json",
func Blocks(c echo.Context) error {

	key := []byte("api_blocks")
	valuesRaw, _ := cache.Get(key)

	if valuesRaw == nil {

		rows, _ := connectClickhouse.Query(`
					SELECT date, countDistinct(block_number)
					FROM Transactions
					GROUP BY date
					ORDER BY date ASC`)
		defer rows.Close()

		var (
			dates       = []string{}
			blockCounts = []int64{}
		)

		for rows.Next() {
			var (
				date       time.Time
				blockCount int64
			)
			if err := rows.Scan(&date, &blockCount); err != nil {
				log.Fatal(err)
				continue
			}

			dates = append(dates, date.Format("Jan _2"))
			blockCounts = append(blockCounts, blockCount)
		}

		values := echo.Map{
			"date":         dates,
			"block_counts": blockCounts,
		}
		valuesMarshal := gotiny.Marshal(&values)
		_ = cache.Set(key, valuesMarshal, 60*5)
		return c.JSON(http.StatusOK, values)
	}

	valuesCache := echo.Map{}
	gotiny.Unmarshal(valuesRaw, &valuesCache)

	return c.JSON(http.StatusOK, valuesCache)
}

//apiv1.GET("/nodes/list.json",

func NodeList(c echo.Context) error {

	key := []byte("index_page_api")
	valuesRaw, _ := cache.Get(key)

	if valuesRaw == nil {
		nodes := []IndexNodePoint{}
		err = connectMysql.Select(&nodes, `SELECT nodes.address, node_type, name, mg_trust, mg_geo, mg_roi, addresses.delegated_amount AS delegated_amount
					FROM nodes
					INNER JOIN addresses ON (nodes.address=addresses.address AND addresses.delegated_amount>= 100000*1e6 AND addresses.delegated_amount <= 10000000*1e6)
					WHERE mg_status=1 AND mg_trust<>'0.001'
					ORDER BY ROUND(delegated_amount/1e11,0) ASC, mg_trust DESC, mg_roi DESC
					LIMIT 500`) // AND mg_roi<>'0.000000'
		if err != nil {
			log.Fatal(err.Error())
		}

		values := echo.Map{
			"nodes": nodes,
		}

		valuesMarshal := gotiny.Marshal(&values)
		_ = cache.Set(key, valuesMarshal, 60)
		return c.JSON(http.StatusOK, values)
	}

	valuesCache := echo.Map{}
	gotiny.Unmarshal(valuesRaw, &valuesCache)

	return c.JSON(http.StatusOK, valuesCache)
}

*/
