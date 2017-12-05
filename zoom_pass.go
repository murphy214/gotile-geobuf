package gotile

import (
	"fmt"
	g "github.com/murphy214/geobuf"
	m "github.com/murphy214/mercantile"
	"database/sql"
	//"math"
	"os"
	"io/ioutil"

)

/*
// calculating memory implications
func Calc_Memory(raw_filesize int,total_features,currentzoom int,maxzoom int) int {
	//gosize := int(math.Pow(float64(maxzoom - currentzoom),2.0)) * 2
	//gosize = gosize * raw_filesize + gosize * 4000
	//return gosize / 1000 // kb
	return 0
}
*/
func Number_Features(currentzoom int,maxzoom int,number_features int) int {
	//fmt.Println((maxzoom - currentzoom),"dif",currentzoom,(maxzoom - currentzoom) * number_features)
	total := 0
	current := currentzoom
	for current < maxzoom {
		number_features = number_features * 3
		total += number_features
		current += 1
	}
	return total
}

// gets the total number of features for each layer
func (filemap *File_Map) Total_Features() int {
	total := 0
	for k,v := range filemap.File_Map {
		total += Number_Features(int(k.Z),filemap.Config.Maxzoom,v.Total_Features)
	}
	return total
}

 




func (filemap *File_Map) Zoom_Pass(db *sql.DB) *sql.DB {
	// iterating through each geobuf
	c := make(chan []Vector_Tile)
	fmt.Println(filemap.Total_Features(),filemap.Config.Currentzoom)
	boolval := false
	if filemap.Total_Features() < 10000000 {
		boolval = true
	}
	for k,v := range filemap.File_Map {
		//memorysize := Calc_Memory(v.File.FileSize,int(k.Z),filemap.Config.Maxzoom)
		go func(k m.TileID,v *g.Geobuf,c chan []Vector_Tile) {
			if boolval == false {
			//fmt.Println(Number_Features(int(k.Z),filemap.Config.Maxzoom,v.Total_Features))
				c <- []Vector_Tile{Make_Tile(k,v,"shit")}
			} else {
				//fmt.Println(memorysize)
				//fmt.Println(memorysize,k)
				bytevals,err := ioutil.ReadFile(v.Filename)
				if nil != err {
					fmt.Println(err)
				}
				first_tile := Make_Tile(k,v,"shit")

				v.File.File.Close()



				os.Remove(v.Filename)
				eh := Make_Zoom_Drill(k,g.Read_FeatureCollection(bytevals),filemap.Config.Prefix,filemap.Config.Maxzoom)
				eh = append(eh,first_tile)


				filemap.SS.Lock()
				delete(filemap.File_Map,k)
				filemap.SS.Unlock()
				c <- eh

			}

		}(k,v,c)
	}

	// collecting tiles
	newlist := []Vector_Tile{}
	count := 0
	size := len(filemap.File_Map)
	for count < size {
		out := <-c
		if len(out) > 1000 {
			Insert_Data3(out,db,filemap.Config)

		} else {
			for _,outt := range out {
				if len(outt.Data) > 0 {
					newlist = append(newlist,outt)
				}
			}
		}

		count += 1
		fmt.Printf("\r[%d/%d] Tiles Made",count,size)
	}
	
	Insert_Data3(newlist,db,filemap.Config)
	return db

}

// creating the tiles
func (filemap *File_Map) Make_Tiles() {
	db := Create_Database_Meta(filemap.Config,filemap.Config.FirstFeature)
	//fmt.Printf("%+v%s",db,"Imheredbdv")
	filemap.Config.Currentzoom = filemap.Zoom
	db = filemap.Zoom_Pass(db)	

	for filemap.Zoom <= filemap.Config.Maxzoom {
		filemap.Config.Currentzoom = filemap.Zoom
		if len(filemap.File_Map) == 0 {
			fmt.Println("shit")
			filemap.Zoom = 10000
		} else {
			fmt.Println("here1")
			filemap = filemap.Drill_Map()
			db = filemap.Zoom_Pass(db)	
		}



		//fmt.Printf("%+v,%+v\n",filemap.Config,filemap.Config.FirstFeature)
		//fmt.Println(vtlist)
		//fmt.Println(filemap.Zoom)
	}




	Make_Index(db)
}

