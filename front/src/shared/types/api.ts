export interface Paginate {
  page: number;
  limit: number;
  total: number;
}

export interface ApiResponse<Data, Item> {
  data: Data | null;
  items: Item[];
  paginate: Paginate | null;
  meta: unknown;
}
