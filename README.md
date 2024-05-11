![image](https://github.com/McFlanky/go-hooktest/assets/153543951/d97ab756-4537-487e-8d36-03facdc0f5ee)

### Why?
Lorem Ipsum

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
   ![image](https://github.com/McFlanky/go-hooktest/assets/153543951/fe82857a-3761-4bff-ad5c-385909a3f99d)
