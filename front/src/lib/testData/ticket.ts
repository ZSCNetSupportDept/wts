import type { Ticket, UserProfile, Trace } from '$lib/types/apiResponse';
import { NowRFC3339, RFC3339 } from '$lib/types/RFC3339';
export const sample1issuer: UserProfile = {
        sid: '2020123456',
        name: '张三',
        block: 'XHA',
        access: 'user',
        room: '123',
        phone: '13800138000',
        isp: 'mobile',
        account: '12345678901@139.gd',
        wx: 'zhangsan_wx',
};

export const sample1: Ticket = {
        tid: 3,
        issuer: sample1issuer,
        description: '网络坏了啊啊啊，快来修！好像是网线被蟑螂咬坏了。',
        occur_at: RFC3339('2023-12-31T23:59:59Z'),
        submitted_at: RFC3339('2024-01-01T00:10:00Z'),
        category: 'ip-or-device',
        status: 'scheduled',
        priority: 'assigned',
        appointed_at: NowRFC3339(),
};

export const sample2issuer: UserProfile = {
        sid: '2020123456',
        name: 'hajimi',
        block: '17',
        access: 'user',
        room: '701',
        phone: '13800138000',
        isp: 'telecom',
        account: '18923456789',
        wx: 'zhangsan_wx',
};

export const sample2: Ticket = {
        tid: 2,
        issuer: sample2issuer,
        description: '才办的宽带，麻烦来装下，谢谢！',
        occur_at: RFC3339('2023-12-31T23:59:59Z'),
        submitted_at: RFC3339('2024-01-01T00:10:00Z'),
        category: 'first-install',
        status: 'fresh',
        priority: 'mainline',
};

export const sampleTrace: Trace[] = [
        {
                opid: 1,
                tid: 10,
                updated_at: RFC3339('2024-01-01T10:00:00Z'),
                op: '2395',
                op_name: '（用户操作）',
                remark: '工单已提交',
                new_status: 'fresh',
                new_priority: 'mainline',
        },
        {
                opid: 2,
                tid: 10,
                updated_at: RFC3339('2024-01-01T10:00:00Z'),
                op: '2395',
                op_name: '哈哈哈',
                remark: '用户预约了时间',
                new_status: 'scheduled',
                new_appointment: RFC3339('2024-01-13T14:00:00Z'),
        },
        {
                opid: 3,
                tid: 10,
                updated_at: RFC3339('2024-01-01T10:00:00Z'),
                op: '2395',
                op_name: '啊啊啊',
                remark: '材料不足，改日修',
                new_status: 'delay',
        },
        {
                opid: 4,
                tid: 10,
                updated_at: RFC3339('2024-01-01T10:00:00Z'),
                op: '2395',
                op_name: '嘿嘿嘿',
                remark: '与用户重新约定时间，预约在2024-01-21下午',
                new_status: 'scheduled',
                new_appointment: RFC3339('2024-01-21T15:00:00Z'),
        },
        {
                opid: 5,
                tid: 10,
                updated_at: RFC3339('2024-01-01T10:00:00Z'),
                op: '2395',
                op_name: '喵喵喵',
                remark: '问题解决：用户路由器线路接触不良，更换后恢复正常',
                new_status: 'solved',
        },
];
