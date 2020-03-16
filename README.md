# go-bitflyer-wrapper  
this pkg is my pkg:go-bitflyer wrapper, for managed.    


2020/03/13: switched go-bitflyer to go-exchanges/api/bitflyer.  
- change private websocket struct  
- change data types & assert type    

## public websocket
### table executions
- set  
- bestask,bestbid  
- ltp  
- spread  
- change price in this term (any any)  


## private websocket
### child orders
- managed orders  
    - has orders  
- managed positions  
    - has size  
- managed cancels  
    - check is canceled  

# Auther
[@numbP](https://twitter.com/_numbp)