export interface Channel {
	id: number;
	name: string;
	projectId: number | null;
}

export interface Member {
	name: string;
	title: string;
	image?: string;
}

export interface UserMessage {
	id: number;
	type: 'USER';
	name: string;
	title: string;
	date: string;
	image?: string;
	message: string;
}

export interface QuoteMessage {
	id: number;
	type: 'QUOTE';
	date: string;
	message: string;
}

export interface InvoiceMessage {
	id: number;
	type: 'INVOICE';
	date: string;
	message: string;
}

export interface FileUploadMessage {
	id: number;
	type: 'FILE_UPLOAD';
	date: string;
	message: string;
	link: string;
}

export type ChannelMessage = UserMessage | QuoteMessage | InvoiceMessage | FileUploadMessage;
