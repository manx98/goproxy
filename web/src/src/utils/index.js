export class Fetcher {
    constructor(reader) {
        this.reader = reader;
        this.done = false;
        this.value = new Uint8Array(0);
    }

    async nextBlock() {
        if (this.done) {
            throw new Error('Already done');
        }
        let {value, done} = await this.reader.read();
        this.done = done;
        if(!done) {
            this.value = value;
        }
    }

    async read(num) {
        let data = new Uint8Array(num);
        for (let i = 0; i < num;) {
            if (this.value.length <= 0) {
                await this.nextBlock();
            } else if (this.value.length + i < num) {
                data.set(this.value, i);
                i += this.value.length;
                this.value = new Uint8Array(0);
            } else {
                data.set(this.value.slice(0, num - i), i);
                this.value = this.value.slice(num - i);
                break
            }
        }
        return data;
    }

    async readByte() {
        const data = await this.read(1)
        return data[0];
    }

    async readInt64() {
        const data = await this.read(8);
        return Number(new DataView(data.buffer).getBigInt64(0, false));
    }

    async readBytes() {
        let num = await this.readInt64();
        if (num <= 0) {
            return new Uint8Array(0);
        }
        return await this.read(num);
    }

    async readString() {
        let data = await this.readBytes()
        if (data.length <= 0) {
            return '';
        } else {
            return new TextDecoder().decode(data);
        }
    }
}

export async function Fetch(url, method = 'GET') {
    const response = await fetch(url, {
        method: method,
    });
    if (response.ok) {
        return new Fetcher(response.body.getReader());
    } else {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
}