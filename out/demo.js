/**
 * GoCall library binding
 * flow
 */

import {
  NativeModules,
  DeviceEventEmitter
} from 'react-native';

async function runApiCall(name: string, args :any[]) : Promise<any> {
    const callData = JSON.stringify({
      args,
      method: name,
    })

    return NativeModules.GoCall.callMethod(callData)
}

/**
 * SimpleCall is good method to test your app
 */
export async function SimpleCall(body: string) : Promise<SimpleResult> {
   try {
        const jsonString = await runApiCall('SimpleCall', [body])
        return JSON.parse(jsonString)
   } catch(error) {
        console.error("Call of SimpleCall failed", error)
        throw error
   }
}

/** 
 * SimpleResult Simple response for your request
 */
export type SimpleResult = {  
    ResponeString: string, 
}
