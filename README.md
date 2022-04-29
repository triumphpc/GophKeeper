# GophKeeper
GophKeeper - it's client-server system which safe and secure save logins and passwords, binary data and other private data.


## Functionality include

### Server
- [x] Docker and project template
- [x] Implement JWT token for client-server auth
- [x] Pg storage
- [x] Data models implementation
- [x] gRPC server-client handlers
- [x] Swagger documentation
#### 
- [ ] (OTP) One-time pass support
- [ ] Message-broker for save user data handlers (RabbitMQ)
- [ ] Secure gRPC connection with SSL/TLS
- [ ] Database encryption
- [ ] Table tests
- [ ] Delete data handlers
- [ ] Edit data handlers
- [ ] Documentation

### Client
- [x] TUI terminal user interface
- [x] Create user data
- [x] Auth by JWT token
####
- [ ] Current build version info
- [ ] View data list
- [ ] Update
- [ ] Delete
- [ ] Search user data
- [ ] Sync clients data 
- [ ] Offline mode
- [ ] Local storage 
- [ ] Windows, Linux, MacOS binary clients
- [ ] Integration tests
- [ ] Documentation