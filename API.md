**Wallet Service API**
----
  Modifies the bank account balances of 2 accounts at a time as a result of a funds transfer via a JSON POST and fetches json data about available multiple bank accounts as well as about the fund transfers between those accounts.

**URL**

  `/accounts`

* **Method:**
  
  GET
  
*  **URL Params**

   None

   **Required:**
 
   None

   **Optional:**
 
   None

* **Data Params**

   None

* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:** `{"v":["Account: alice456  Balance = 573.810000 USD  Initial Balance = 573.810000","Account: bob123  Balance = 302.350000 USD  Initial Balance = 302.350000","Account: lucy0123  Balance = 14583.900000 EUR  Initial Balance = 14583.900000","Account: marcy789  Balance = 4583.900000 EUR  Initial Balance = 4583.900000"]}`
 
* **Error Response:**

  * **Code:** 200 <br />
    **Content:** `{"v":null,"err":"err: error begining transaction in postgresdial tcp 127.0.0.1:5432: connect: connection refused"}`

    OR

  * **Code:** 200 <br />
    **Content:** `{"v":null,"err":"err: error begining transaction in postgrespq: sorry, too many clients already"}`
    
    OR

  * Failed to connect to 127.0.0.1 port 8080: Connection refused
  
    OR

  * **Code:** 404 <br />
    **Content:** `404 page not found`
    

* **Sample Call:**

  ```curl -i "127.0.0.1:8080/accounts"```


**URL**

  `/submittransfer`

* **Method:**
  
  `POST`
  
*  **URL Params**

   None

   **Required:**
 
   None

   **Optional:**
 
   None

* **Data Params**

  `{"from":"bob123","to":"alice456","amount":"20"}`

* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:** `{"result":"success"}`
 
* **Error Response:**

  * **Code:** 200 <br />
    **Content:** `{"v":null,"err":"err: error begining transaction in postgresdial tcp 127.0.0.1:5432: connect: connection refused"}`

    OR

  * **Code:** 200 <br />
    **Content:** `{"v":null,"err":"err: error begining transaction in postgrespq: sorry, too many clients already"}`
    
    OR

  * `Failed to connect to 127.0.0.1 port 8080: Connection refused`
  
    OR

  * **Code:** 404 <br />
    **Content:** `404 page not found`
    

* **Sample Call:**

  ```curl  -d'{"from":"bob123","to":"alice456","amount":"20"}' "0.0.0.0:8080/submittransfer"```

**URL**

 `/transfers`

* **Method:**
  
  GET
  
*  **URL Params**

   None

   **Required:**
 
   None

   **Optional:**
 
   None

* **Data Params**

  None

* **Success Response:**
  
  * **Code:** 200 <br />
    **Content:** `{"v":["Transfer #1  from: bob123  to:  alice456 in the amount of 20.000000 USD at 2019-03-25T12:02:55Z"]}`
 
* **Error Response:**


  * **Code:** 200 <br />
    **Content:** `{"v":null,"err":"err: error begining transaction in postgresdial tcp 127.0.0.1:5432: connect: connection refused"}`

    OR

  * **Code:** 200 <br />
    **Content:** `{"v":null,"err":"err: error begining transaction in postgrespq: sorry, too many clients already"}`
    
    OR

  * `Failed to connect to 127.0.0.1 port 8080: Connection refused`
  
    OR

  * **Code:** 404 <br />
    **Content:** `404 page not found`
    

* **Sample Call:**

  ```curl -i "127.0.0.1:8080/tarnsfers"```


