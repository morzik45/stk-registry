import http from "../http-common";

class BreakersDataService {
    getAll() {
        return http.get("/breakers");
    }
    check(snils, checked) {
        console.log(snils, checked);
        return http.post(`/breakers/check?snils=${snils}&checked=${checked}`);
    }
    saveToExcel(data) {
        return http.post(`/breakers/make-excel`, JSON.stringify(data), { responseType: 'arraybuffer' })
    }
}

export default new BreakersDataService();