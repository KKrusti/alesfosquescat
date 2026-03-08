export interface StatsResponse {
  total_this_year: number
  longest_incident_streak: number
  days_since_last_incident: number
  last_incident_date: string
  longest_no_incident_streak: number
  current_incident_streak: number
}

export type AnimState =
  | 'idle'
  | 'loading'
  | 'success'
  | 'already_voted'
  | 'network_error'
  | 'server_error'
