# reakgo
Simple Framework to quickly build webapps in GoLang

# Backend Supports both Golang POST and Curl requests

## Golang POST Endpoints
* APPURL/login: 	serve login page and on submit serve dashboard page on valid credentials.
* APPURL/signup: 	serve signup page and on submit create new account.
* APPURL/verifyEmail: 	verify email if clicked on valid link.
* APPURL/forgotPassword:	serve forgotPassword page and on submit send email with link to change password if provide registered email. 
* APPURL/changePassword:	on submit change password if clicked on valid change password link.

# Authenciation Curl API's

## r.Method
* "GET": 	getting single or multiple records
* "POST": 	update single or multiple records
* "PUT": 	insert single or multiple records
* "DELETE":	delete single or multiple records	 


### PUT user (sing-in)
```
curl -X PUT http://localhost:4001/signup -H 'reak-api:true' -d '{"Email":"anamika.namdeo@reak.in","Password":"123", "Type":"admin" , "FirstName":"nitin", "LastName":"nitin"}'
```
{"Status":"success","Message":"Congratulations, you have successfully enrolled","Payload":{"Token":"vMBPOMwVVNg9br1EQU7pOWMOuM42TJnud5YsvT2jIwVRtoBA-0fwfk1U7OV7"}}


// tried with same email again
```
curl -X PUT http://localhost:4001/signup -H 'reak-api:true' -d '{"Email":"anamika.namdeo@reak.in","Password":"123", "Type":"user" , "FirstName":"nitin", "LastName":"nitin"}'
```
{"Status":"failure","Message":" Duplicate entry 'anamika.namdeo@reak.in' for key 'authentication.email'","Payload":null}

// without required fields

```
curl -X PUT http://localhost:4001/signup -H 'reak-api:true' -d '{"Email":"","Password":"123", "Type":"user" , "FirstName":"nitin", "LastName":"nitin"}'
```
{"Status":"failure","Message":"Please fill all required fields and try again","Payload":null}


### verify email

```http://localhost:4001/verifyEmail?email=anamika.namdeo@reak.in&token=udh05SyVpOs9yQtgTqHiSjcARwtUuM99cRXScF6b-fxXtEdGUz9vOYMcurbZ```

{"Status":"success","Message":"Email Verification was successful, Please continue by logging in","Payload":null}

// clicked again
{"Status":"success","Message":"Please login to continue using the application","Payload":null}

//after 30 minutes
{"Status":"failure","Message":"Email verification link was expired. Please contact the administrator atsupport@reak.in","Payload":null}

// email invalid
```http://localhost:4001/verifyEmail?email=anamika.na@reak.in&token=udh05SyVpOs9yQtgTqHiSjcARwtUuM99cRXScF6b-fxXtEdGUz9vOYMcurbZ```

{"Status":"failure","Message":"Authentication Failure, please click on valid Email","Payload":null}

// token invalid
```http://localhost:4001/verifyEmail?email=anamika.namdeo@reak.in&token=udh05SyVpOs9yQtgTqHiSjcARwtUuM99cRXScF6b-fxEdGUz9vOYMcurbZ```

{"Status":"failure","Message":"Authentication Failure, please click on valid Email","Payload":null}

//email or token missing
```http://localhost:4001/verifyEmail?email=anamika.namdeo@reak.in```

{"Status":"failure","Message":"Unable to verify the email as it lacks the required values. Please retry and if the problem persists reach out to the administrator at support@reak.in","Payload":null}


## login
```
curl -X POST http://localhost:4001/login -H 'reak-api:true' -d '{"Email":"anamika.namdeo@reak.in","Password":"123"}'
```

{"Status":"info","Message":"Please login in the website to continue","Payload":null}

//with both empty
```
curl -X POST http://localhost:4001/login -H 'reak-api:true' -d '{"Email":"","Password":""}'
```

{"Status":"failure","Message":"Please fill all required fields and try again","Payload":null}

// with notRegistered/invalid email
```
curl -X POST http://localhost:4001/login -H 'reak-api:true' -d '{"Email":"test@yop","Password":"123"}'
```

{"Status":"failure","Message":"Incorrect Credentials, Please re-check and try again","Payload":null}

// with valid email but wrong password
```
curl -X POST http://localhost:4001/login -H 'reak-api:true' -d '{"Email":"anamika.namdeo@reak.in","Password":"12378"}'
```

{"Status":"failure","Message":"Incorrect Credentials, Please re-check and try again","Payload":null}




## forgotPassword

```
curl -X POST http://localhost:4001/forgotPassword -H 'reak-api:true' -d '{"Email":"anamika.namdeo@reak.in"}'
```
{"Status":"success","Message":"New password has been sent to you via email, Please check your inbox","Payload":null}

// invalid/not registered email
```
curl -X POST http://localhost:4001/forgotPassword -H 'reak-api:true' -d '{"Email":"test@yop"}'
```
{"Status":"success","Message":"New password has been sent to you via email, Please check your inbox","Payload":null}













