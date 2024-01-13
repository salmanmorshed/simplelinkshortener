export type Link = {
    slug: string;
    url: string;
    visits: string;
    created_at: string;
};

export type NewLink = {
    short_url: string;
};

export type Paginated<T> = {
    results: T[];
    total: number;
    limit: number;
    offset: number;
};
