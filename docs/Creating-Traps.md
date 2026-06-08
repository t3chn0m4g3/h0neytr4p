## What is a trap?

A trap is a `.json` file which consists of behaviour related a trap. For example, you can define that it should trap any request that matches `/jenkins`. All requests to `/jenkins` will be captured.

## Structure of a trap
Each trap must contain the following details in the JSON format. Most of the elements are self explanatory. A summarised version of all supported elements are mentioned below.

|S.No | Name of the attribute  | Description of the attribute  | Required/Optional  |  Data Type |
|---|---|---|---|---|
| 1.  |  BasicInfo | This is a nested json containing the basic information of the trap  | Required  |  JSON object |
| 2.  |  BasicInfo.Name | This is present inside basic info wrapper. This contains the name of the trap. Has to be unique across all the traps for better detection  |  Required | String  |
| 3.  | BasicInfo.Port  | The port where the trap listener should be started.  | Required   | String  |
| 4.  | BasicInfo.Protocol  |  The protocol family for the trap, for example `HTTP` or `HTTP/2`.  |  Required |  String |
| 5.  | BasicInfo.MitreAttackTags  |  For now, this will have external facing web compromise technique ID. However, it's reserved for future traps which will support more protocols |  Required | Strings separated with a comma  |
| 6.  | BasicInfo.References  | Any reference URL for the attack for better detection and analytics.  | Required  | String  |
| 7.  |  BasicInfo.Description | Description of the attack which is being trapped.  | Required  | String  |
| 8.  |  BasicInfo.RiskRating | RiskRating for the trap; No defined values yet but generally accepted: Critical, High, Medium, Low, Info  | Required  | String  |
| 9.  |  Behaviour | It's a json object containing a pair of request and response.  | Required  | JSON array  |
| 10.  | Behaviour.Request  |  It's a json object containing the required request behaviour | Required  | JSON Object  |
| 11.  |  Behaviour.Request.Url |  URL Path.  You can define static or wildcards. [See reference below for how-to] | Required  | String  |
| 12.  |  Behaviour.Request.Method | Method; Supported Values: GET, POST, DELETE, PUT, OPTIONS  |  Required |  String |
| 13.  | Behaviour.Request.Proto | Optional HTTP request protocol matcher, for example `HTTP/2*`. Omit it or use an empty string to match any HTTP version. | Optional | String |
| 14.  | Behaviour.Request.Headers  | Request Headers. You can define static or wildcard values. [See reference below for how-to]; Use an empty `{}` for empty headers | Required  |  JSON Object |
| 15.  |  Behaviour.Request.Params |  It's a key value pair json object. You can define the parameters irrespective of GET/POST and Content-Types. Backend handles it automatically. Use an empty `{}` for empty parameters. | Required  | JSON Object  |
| 16.  |  Behaviour.Response |  It's a json object containing the required response behaviour | Required  | JSON Object  |
| 17.  |  Behaviour.Response.StatusCode | Status code for response;  |  Required | String  |
| 18.  |  Behaviour.Response.Body | File content or location of the file;   |   |   |
| 19.  |  Behaviour.Response.Type | file or string  | Required | String  |
| 20.  |  Behaviour.trap | "true" if you want to trap, "false" if you don't want to   | Required  | String  |


#### Example Trap:
```
{
    "BasicInfo": {
        "Name":"jenkins_home",
        "Port":"443",
        "Protocol":"HTTP",
        "MitreAttackTags":"",
        "References":"",
        "RiskRating":"Critical",
        "Description":""
    },
    "Behaviour":
    [
        {
            "Request": {
                "Url":"/jenkins*",
                "Method": "GET",
                "Proto":"",
                "Headers":{"User-Agent":"*"},
                "Params":{}
            },
            "Response": {
                "StatusCode": 302,
                "Body": "traps/assets/jenkins/default.html",
                "Type":"file"
            },
            "trap": "true"
        }
    ]
}
```

### Defining pattern inside an attribute:

Most request matchers use glob-style `*` wildcards.

###### Quick Walkthrough:

`*` - wildcard

Basically, 
- Let's say you want to match `/jenkins` in the Url field, you will use `/jenkins`. You can use `*` for defining a wildcard entry.
- Let's say you want to match `/wp-admin/login` in the Url field, you will use `/wp-admin/login`. 
- Let's say you want to match `/login.php?id=1'` and `/login.php?id=<ANY_Number>'`, use `"Url":"/login.php"` and `"Params":{"id":"*'"}`.

The same goes for headers and params.

Use `"Proto":"HTTP/2*"` when a trap should only match HTTP/2 requests. Leave `Proto` empty or omit it when the HTTP version does not matter.

More examples: 

- You want to create a header list which accepts anything that starts with mozilla.: `"Headers": {"User-Agent":"Mozilla*"}` will be your header value.
- You want to match decoded HTTP Basic authentication content without logging decoded credentials: `"Headers": {"Authorization-Basic-Decoded":"*known-marker*"}`. This is a virtual matcher, not a real request header.
- You want to create a parameter set which accepts username: anything starting with admin and password: password only.: `"Params":{"username":"*admin*","password":"password"}`


### Isn't there an automated way of converting requests into traps? 
Coming soon.

### Is there a way to verify if my trap syntax is right?
You can use any JSON lint tool. Specific syntax based checking is coming soon.
