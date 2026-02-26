// 用于表单提交时的校验方便
export class invalidState {
        notOK = $state(false);
        txt = $state('');
        assert(x: boolean, e: string) {
                if (!x && !this.notOK) {
                        this.notOK = true;
                        this.txt = e;
                }
        }

        reset() {
                this.notOK = false;
                this.txt = '';
        }
}