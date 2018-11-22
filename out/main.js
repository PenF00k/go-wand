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
 * ParseFile - Parse file
 */
export async function Parse(src: string, output: string, goutput: string) : Promise<any> {
   try {
        const jsonString = await runApiCall('Parse', [src, output, goutput])
        return JSON.parse(jsonString)
   } catch(error) {
        console.error("Call of Parse failed", error)
        throw error
   }
}

