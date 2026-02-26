import type { FilterTicketsReq } from '$lib/types/apiRequest';
import type { WtsZone } from '$lib/types/enum';

export type Criteria = {
        r: FilterTicketsReq;
        _order: 'priority' | 'newest' | 'oldest';
        _floor: number | null;
        _blocks_in_zone: WtsZone[]; //需要和req.block保持一致，注意
        _view_today_scheduled: boolean;
};

export let criteria: Criteria = {
        r: {
                scope: 'active',
                issuer: undefined,
                block: [],
                status: [],
                priority: [],
                category: [],
                isp: [],
                newer_than: undefined,
                older_than: undefined,
        },
        _order: 'priority',
        _floor: null,
        _blocks_in_zone: [],
        _view_today_scheduled: false
} as Criteria;

export function resetCriteria() {
        criteria = {
                r: {
                        scope: 'active',
                        issuer: undefined,
                        block: [],
                        status: [],
                        priority: [],
                        category: [],
                        isp: [],
                        newer_than: undefined,
                        older_than: undefined,
                },
                _order: 'priority',
                _floor: null,
                _blocks_in_zone: [],
                _view_today_scheduled: false
        } as Criteria;

}
