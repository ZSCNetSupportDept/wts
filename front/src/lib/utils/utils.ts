export function IsAssigned(issuer: string) : boolean {
        const assigned = ['gd1','gd2','gd3','gd4','gd5','gd6','gd7','gd8','gd9','gd10','gd11','gd12','gd13','gd14','gd15','gd16','gd17','gd18','gd19','gd20','gd21','gd22','gdXHA','gdXHB','gdXHC','gdXHD','gdZH','gdOther']
        return assigned.includes(issuer)
}