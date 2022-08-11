import http from "../http-common";

class BreakersDataService {
    getAll() {
        return http.get("/breakers/");
    }
    check(snils) {
        return http.post(`/breakers/check/?snils=${snils}`);
    }
    saveToExcel(data) {
        return http.post(`/breakers/make-excel/`, JSON.stringify(data), { responseType: 'arraybuffer' })
    }
}

export default new BreakersDataService();