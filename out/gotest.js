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
 * CallAndGet use for your classes
 */
export async function CallAndGet(id: string, params: number[]) : Promise<LovelyStructure> {
   try {
        const jsonString = await runApiCall('CallAndGet', [id, params])
        return JSON.parse(jsonString)
   } catch(error) {
        console.error("Call of CallAndGet failed", error)
        throw error
   }
}


/**
 * Receive change notification for a stage with ID
 */
export async function StageSubscription(id: number) {
   try {
      await subribeApiCall('StageSubscription', [id])
   } catch(error) {
      console.error("Call of StageSubscription failed", error)
   }
}

export async function cancelStageSubscription(id: number) {
   try {
      await cancelSubsriptionApiCall('StageSubscription', [id])
   } catch(error) {
      console.error("Call of StageSubscription failed", error)
   }
}

/** 
 * LovelyInterface My Comment
 */
export type Simple = {  
    SuName: string, 
}
/** 
 * LovelyStructure for your code
 * Test my code
 */
export type LovelyStructure = {  
    // Type name test 
    TypeName: string,  
    // Optional name test 
    TypeOptName: ?string,  
    // number test 
    Number: number,  
    // list test 
    NumberList: number[],  
    // Map example 
    Map: { [key: string]: { [key: string]: Simple}[]}, 
}
