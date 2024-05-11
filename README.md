![image](https://github.com/McFlanky/go-hooktest/assets/153543951/d97ab756-4537-487e-8d36-03facdc0f5ee)

### Why?
To help developers test their webhooks from using third party apis like Stripe all on the CLI...

### Overview
![image](https://github.com/McFlanky/go-hooktest/assets/153543951/c65ca380-4892-4dd8-951b-f091f3a75ddf)

### How To Run:
1) Create Host Keys
   ```
   sudo chmod +x keygen.sh
   ./keygen.sh
   ```
2) Run the Server:
   ```
   make
   ```
3) How to Use:
   ```
   ssh localhost -p 2222
   ```
4) Enter Webhook Destination in Form of:
   ```
   http://localhost:3000
   ```
   Then it generates a random webhook on local host 5000 w/ unique ID (shown below):
   ![image](https://github.com/McFlanky/go-hooktest/assets/153543951/af332a6c-e402-4bdf-b360-016dcbe70b6b)

   
   You then Copy & Run Command to tunnel the traffic to given random generated port (shown below):
   ![image](https://github.com/McFlanky/go-hooktest/assets/153543951/5fa4a1ef-73be-4288-8fda-0ba211218962)

