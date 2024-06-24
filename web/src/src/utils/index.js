const DirAddUpdated = 68;
const SizeAddUpdated = 83;
const StatusUpdated = 77;
const BinaryWrite = 87;

export class Fetcher {
    constructor(reader) {
        this.reader = reader;
        this.done = false;
        this.value = new Uint8Array(0);
        this.onDirAddUpdated = (val) => {
        }
        this.onSizeAddUpdated = (val) => {
        }
        this.onStatusUpdated = (val) => {
        }
        this.onBinaryWrite = (val) => {
        }
    }

    async nextBlock() {
        if (this.done) {
            throw new Error('Already done');
        }
        let data = await this.reader.read();
        this.done = data.done;
        if (!data.done) {
            this.value = data.value;
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

    async run() {
        try {
            while (true) {
                let t = await this.readByte();
                switch (t) {
                    case DirAddUpdated:
                        await this.onDirAddUpdated(await this.readInt64());
                        break;
                    case SizeAddUpdated:
                        await this.onSizeAddUpdated(await this.readInt64());
                        break;
                    case StatusUpdated:
                        await this.onStatusUpdated(await this.readString());
                        return;
                    case BinaryWrite:
                        await this.onBinaryWrite(await this.readBytes());
                        break;
                    default:
                        throw new Error(`Unknown type ${t}`);
                }
            }
        } finally {
            await this.reader.cancel()
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

export function formatBytes(bytes, decimals = 2) {
    if (bytes === undefined || bytes === 0) return '0B';

    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + sizes[i];
}

export function formatTime(timestamp) {
    const date = new Date(timestamp);

    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0'); // 月份从0开始，需要加1
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${year}/${month}/${day} ${hours}:${minutes}`;
}

export function getDiffZipFileName() {
    const date = new Date();
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0'); // 月份从0开始，需要加1
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${year}_${month}_${day}_${hours}_${minutes}_diff.zip`;
}