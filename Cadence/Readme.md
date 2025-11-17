
* Heartbeat in long running activities:
  * Activities that run for a long time should periodically send heartbeats to Cadence to indicate that they are still alive. This helps Cadence to detect and handle activity failures.
  * Heartbeats can be sent using the `activity.RecordHeartbeat` function provided by the Cadence client library.
  * Heartbeat Payload saves progress for retries
  * Why “short heartbeat timeout” helps: If your activity heartbeats every 10 seconds and the worker dies:
    * Cadence will notice within ~10 seconds. The activity will be retried on another worker quickly. 
    * If you never heartbeat (or have a 2-hour timeout). Cadence might wait 2 hours before retrying — very slow recovery.