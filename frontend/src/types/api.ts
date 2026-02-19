export interface ApiResponse<T> {
  data: T
}

export interface PagedData<T> {
  items: T[]
  total: number
  page: number
  limit: number
}
